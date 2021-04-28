// Copyright 2021 SpecializedGeneralist
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server_test

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/grpcapi"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/hnswgo"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/indexmanager"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/server"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"os"
	"path"
	"testing"
)

func TestServer_CreateIndex(t *testing.T) {
	t.Parallel()

	t.Run("successful creation", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		_, err := srv.CreateIndex(ctx, sampleCreateIndexRequest)
		assert.NoError(t, err)
		assert.Equal(t, []string{"foo"}, im.IndicesNames())
	})

	t.Run("invalid index space", func(t *testing.T) {
		t.Parallel()
		im := indexmanager.New(os.TempDir(), zerolog.Nop())
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		_, err := srv.CreateIndex(ctx, &grpcapi.CreateIndexRequest{
			IndexName:      "foo",
			Dim:            5,
			EfConstruction: 200,
			M:              10,
			MaxElements:    10,
			Seed:           100,
			SpaceType:      3,
			AutoId:         false,
		})
		assert.Error(t, err)
		assert.Empty(t, im.IndicesNames())
	})

	t.Run("creation error", func(t *testing.T) {
		t.Parallel()
		im := indexmanager.New(os.TempDir(), zerolog.Nop())
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		_, err := srv.CreateIndex(ctx, &grpcapi.CreateIndexRequest{
			IndexName:      "foo?!", // invalid name
			Dim:            5,
			EfConstruction: 200,
			M:              10,
			MaxElements:    10,
			Seed:           100,
			SpaceType:      grpcapi.CreateIndexRequest_COSINE,
			AutoId:         false,
		})
		assert.Error(t, err)
		assert.Empty(t, im.IndicesNames())
	})
}

func TestServer_DeleteIndex(t *testing.T) {
	t.Parallel()

	t.Run("successful deletion", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := createManagerWithPersistedIndices(t, dir)
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		assert.Contains(t, im.IndicesNames(), "test-index-auto-id-1")

		_, err := srv.DeleteIndex(ctx, &grpcapi.DeleteIndexRequest{
			IndexName: "test-index-auto-id-1",
		})
		assert.Nil(t, err)

		assert.NotContains(t, im.IndicesNames(), "test-index-auto-id-1")
	})

	t.Run("deletion error", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		_, err := srv.DeleteIndex(ctx, &grpcapi.DeleteIndexRequest{
			IndexName: "foo",
		})
		assert.Error(t, err)
	})
}

func TestServer_InsertVector(t *testing.T) {
	t.Parallel()

	t.Run("successful insertion", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := createManagerWithPersistedIndices(t, dir)
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		resp, err := srv.InsertVector(ctx, &grpcapi.InsertVectorRequest{
			IndexName: "test-index-auto-id-1",
			Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
		})
		assert.NoError(t, err)
		assert.Equal(t, "3", resp.Id)

		assertIndexContainsExactlyIDs(t, im, "test-index-auto-id-1", []uint32{1, 2, 3})
	})

	t.Run("insertion error", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := createManagerWithPersistedIndices(t, dir)
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		resp, err := srv.InsertVector(ctx, &grpcapi.InsertVectorRequest{
			IndexName: "test-index-custom-id-1",
			Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("index not found", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		resp, err := srv.InsertVector(ctx, &grpcapi.InsertVectorRequest{
			IndexName: "foo",
			Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestServer_InsertVectorWithId(t *testing.T) {
	t.Parallel()

	t.Run("successful insertion", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := createManagerWithPersistedIndices(t, dir)
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		resp, err := srv.InsertVectorWithId(ctx, &grpcapi.InsertVectorWithIdRequest{
			IndexName: "test-index-custom-id-1",
			Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
			Id:        42,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assertIndexContainsExactlyIDs(t, im, "test-index-custom-id-1", []uint32{1, 2, 42})
	})

	t.Run("insertion error", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := createManagerWithPersistedIndices(t, dir)
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		resp, err := srv.InsertVectorWithId(ctx, &grpcapi.InsertVectorWithIdRequest{
			IndexName: "test-index-auto-id-1",
			Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
			Id:        42,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("index not found", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		resp, err := srv.InsertVectorWithId(ctx, &grpcapi.InsertVectorWithIdRequest{
			IndexName: "foo",
			Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
			Id:        42,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestServer_InsertVectors(t *testing.T) {
	t.Parallel()

	t.Run("successful insertion", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		{
			im := createManagerWithPersistedIndices(t, dir)
			srv := server.New(sampleServerConfig, im, zerolog.Nop())

			stream := newInsertVectorsServerStream([]*grpcapi.InsertVectorRequest{
				{
					IndexName: "test-index-auto-id-1",
					Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
				},
				{
					IndexName: "test-index-auto-id-1",
					Vector:    &grpcapi.Vector{Value: sampleVectors[1]},
				},
				{
					IndexName: "test-index-auto-id-2",
					Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
				},
			})
			err := srv.InsertVectors(stream)
			assert.NoError(t, err)
			assert.NotNil(t, stream.Reply)
			assert.Equal(t, []string{"3", "4", "3"}, stream.Reply.Ids)
		}
		{
			// Ensure the index was persisted
			im := indexmanager.New(dir, zerolog.Nop())
			err := im.LoadIndices()
			assert.NoError(t, err)

			assertIndexContainsExactlyIDs(t, im, "test-index-auto-id-1", []uint32{1, 2, 3, 4})
			assertIndexContainsExactlyIDs(t, im, "test-index-auto-id-2", []uint32{1, 2, 3})
		}
	})

	t.Run("insertion error", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		im := createManagerWithPersistedIndices(t, dir)
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		stream := newInsertVectorsServerStream([]*grpcapi.InsertVectorRequest{
			{
				IndexName: "test-index-custom-id-1",
				Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
			},
		})
		err := srv.InsertVectors(stream)
		assert.Error(t, err)
		assert.Nil(t, stream.Reply)
	})

	t.Run("index not found", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		stream := newInsertVectorsServerStream([]*grpcapi.InsertVectorRequest{
			{
				IndexName: "foo",
				Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
			},
		})
		err := srv.InsertVectors(stream)
		assert.Error(t, err)
		assert.Nil(t, stream.Reply)
	})
}

func TestServer_InsertVectorsWithIds(t *testing.T) {
	t.Parallel()

	t.Run("successful insertion", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		{
			im := createManagerWithPersistedIndices(t, dir)
			srv := server.New(sampleServerConfig, im, zerolog.Nop())

			stream := newInsertVectorsWithIdsServerStream([]*grpcapi.InsertVectorWithIdRequest{
				{
					IndexName: "test-index-custom-id-1",
					Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
					Id:        10,
				},
				{
					IndexName: "test-index-custom-id-1",
					Vector:    &grpcapi.Vector{Value: sampleVectors[1]},
					Id:        11,
				},
				{
					IndexName: "test-index-custom-id-2",
					Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
					Id:        12,
				},
			})
			err := srv.InsertVectorsWithIds(stream)
			assert.NoError(t, err)
			assert.NotNil(t, stream.Reply)
		}
		{
			// Ensure the index was persisted
			im := indexmanager.New(dir, zerolog.Nop())
			err := im.LoadIndices()
			assert.NoError(t, err)

			assertIndexContainsExactlyIDs(t, im, "test-index-custom-id-1", []uint32{1, 2, 10, 11})
			assertIndexContainsExactlyIDs(t, im, "test-index-custom-id-2", []uint32{1, 2, 12})
		}
	})

	t.Run("insertion error", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		im := createManagerWithPersistedIndices(t, dir)
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		stream := newInsertVectorsWithIdsServerStream([]*grpcapi.InsertVectorWithIdRequest{
			{
				IndexName: "test-index-auto-id-1",
				Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
				Id:        10,
			},
		})
		err := srv.InsertVectorsWithIds(stream)
		assert.Error(t, err)
		assert.Nil(t, stream.Reply)
	})

	t.Run("index not found", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		stream := newInsertVectorsWithIdsServerStream([]*grpcapi.InsertVectorWithIdRequest{
			{
				IndexName: "foo",
				Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
				Id:        10,
			},
		})
		err := srv.InsertVectorsWithIds(stream)
		assert.Error(t, err)
		assert.Nil(t, stream.Reply)
	})
}

func TestServer_SearchKNN(t *testing.T) {
	t.Parallel()

	t.Run("successful search", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := createManagerWithPersistedIndices(t, dir)
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		resp, err := srv.SearchKNN(ctx, &grpcapi.SearchRequest{
			IndexName: "test-index-auto-id-1",
			Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
			K:         2,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.Len(t, resp.Hits, 2)
		assert.Equal(t, "1", resp.Hits[0].Id)
		assert.InDelta(t, 0.0, resp.Hits[0].Distance, 1e-6)
		assert.Equal(t, "2", resp.Hits[1].Id)
		assert.Greater(t, resp.Hits[1].Distance, resp.Hits[0].Distance)
	})

	t.Run("index not found", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		resp, err := srv.SearchKNN(ctx, &grpcapi.SearchRequest{
			IndexName: "foo",
			Vector:    &grpcapi.Vector{Value: sampleVectors[0]},
			K:         10,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestServer_FlushIndex(t *testing.T) {
	t.Parallel()

	t.Run("successful flush", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		_, err := im.CreateIndex("foo", hnswgo.Config{
			SpaceType:      hnswgo.CosineSpace,
			Dim:            5,
			MaxElements:    10,
			M:              10,
			EfConstruction: 200,
			RandSeed:       100,
			AutoIDEnabled:  true,
		})
		assert.NoError(t, err)

		err = os.Remove(path.Join(dir, "foo", "index"))
		assert.NoError(t, err)

		resp, err := srv.FlushIndex(ctx, &grpcapi.FlushRequest{
			IndexName: "foo",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.FileExists(t, path.Join(dir, "foo", "index"))
	})

	t.Run("request error", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		resp, err := srv.FlushIndex(ctx, &grpcapi.FlushRequest{
			IndexName: "foo",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestServer_Indices(t *testing.T) {
	t.Parallel()
	dir := createTempDir(t)
	defer deleteDir(t, dir)
	im := indexmanager.New(dir, zerolog.Nop())
	srv := server.New(sampleServerConfig, im, zerolog.Nop())

	cfg := hnswgo.Config{
		SpaceType:      hnswgo.CosineSpace,
		Dim:            5,
		MaxElements:    10,
		M:              10,
		EfConstruction: 200,
		RandSeed:       100,
		AutoIDEnabled:  true,
	}

	resp, err := srv.Indices(ctx, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Indices)

	_, err = im.CreateIndex("foo", cfg)
	assert.NoError(t, err)

	resp, err = srv.Indices(ctx, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, []string{"foo"}, resp.Indices)

	_, err = im.CreateIndex("bar", cfg)
	assert.NoError(t, err)

	resp, err = srv.Indices(ctx, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Indices, 2)
	assert.Contains(t, resp.Indices, "foo")
	assert.Contains(t, resp.Indices, "bar")
}

func TestServer_SetEf(t *testing.T) {
	t.Parallel()

	t.Run("successful set", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := createManagerWithPersistedIndices(t, dir)
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		resp, err := srv.SetEf(ctx, &grpcapi.SetEfRequest{
			IndexName: "test-index-auto-id-1",
			Value:     42,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("index not found", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())
		srv := server.New(sampleServerConfig, im, zerolog.Nop())

		resp, err := srv.SetEf(ctx, &grpcapi.SetEfRequest{
			IndexName: "foo",
			Value:     42,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

var (
	sampleCreateIndexRequest = &grpcapi.CreateIndexRequest{
		IndexName:      "foo",
		Dim:            5,
		EfConstruction: 200,
		M:              10,
		MaxElements:    10,
		Seed:           100,
		SpaceType:      grpcapi.CreateIndexRequest_COSINE,
		AutoId:         false,
	}
	sampleServerConfig = server.Config{
		Address:    "0.0.0.0:0",
		TLSEnabled: false,
		TLSCert:    "",
		TLSKey:     "",
	}
	sampleVectors = [][]float32{
		{0.1, 0.2, 0.3, 0.4, 0.5},
		{0.9, 0.8, 0.7, 0.6, 0.5},
	}
	ctx = context.Background()
)

func createManagerWithPersistedIndices(t *testing.T, path string) *indexmanager.IndexManager {
	t.Helper()
	im := indexmanager.New(path, zerolog.Nop())

	for _, autoID := range []bool{true, false} {
		var namePrefix string
		if autoID {
			namePrefix = "test-index-auto-id"
		} else {
			namePrefix = "test-index-custom-id"
		}

		for i := 1; i <= 2; i++ {
			name := fmt.Sprintf("%s-%d", namePrefix, i)
			index, err := im.CreateIndex(name, hnswgo.Config{
				SpaceType:      hnswgo.CosineSpace,
				Dim:            5,
				MaxElements:    10,
				M:              10,
				EfConstruction: 200,
				RandSeed:       100,
				AutoIDEnabled:  autoID,
			})
			assert.NoError(t, err)

			if autoID {
				for i, vector := range sampleVectors {
					id, err := index.AddPointAutoID(vector)
					assert.NoError(t, err)
					assert.Equal(t, i+1, int(id))
				}
			} else {
				for i, vector := range sampleVectors {
					err = index.AddPoint(vector, uint32(i+1))
					assert.NoError(t, err)
				}
			}

			err = im.PersistIndex(name)
			assert.NoError(t, err)
		}
	}

	return im
}

type insertVectorsServerStream struct {
	baseVectorsStream
	reqIndex int
	Requests []*grpcapi.InsertVectorRequest
	Reply    *grpcapi.InsertVectorsReply
}

var _ grpcapi.Server_InsertVectorsServer = &insertVectorsServerStream{}

func newInsertVectorsServerStream(requests []*grpcapi.InsertVectorRequest) *insertVectorsServerStream {
	return &insertVectorsServerStream{Requests: requests}
}

func (s *insertVectorsServerStream) SendAndClose(reply *grpcapi.InsertVectorsReply) error {
	s.Reply = reply
	return nil
}

func (s *insertVectorsServerStream) Recv() (*grpcapi.InsertVectorRequest, error) {
	if s.reqIndex >= len(s.Requests) {
		return nil, io.EOF
	}
	req := s.Requests[s.reqIndex]
	s.reqIndex++
	return req, nil
}

type insertVectorsWithIdsServerStream struct {
	baseVectorsStream
	reqIndex int
	Requests []*grpcapi.InsertVectorWithIdRequest
	Reply    *grpcapi.InsertVectorsWithIdsReply
}

var _ grpcapi.Server_InsertVectorsWithIdsServer = &insertVectorsWithIdsServerStream{}

func newInsertVectorsWithIdsServerStream(requests []*grpcapi.InsertVectorWithIdRequest) *insertVectorsWithIdsServerStream {
	return &insertVectorsWithIdsServerStream{Requests: requests}
}
func (s *insertVectorsWithIdsServerStream) SendAndClose(reply *grpcapi.InsertVectorsWithIdsReply) error {
	s.Reply = reply
	return nil
}

func (s *insertVectorsWithIdsServerStream) Recv() (*grpcapi.InsertVectorWithIdRequest, error) {
	if s.reqIndex >= len(s.Requests) {
		return nil, io.EOF
	}
	req := s.Requests[s.reqIndex]
	s.reqIndex++
	return req, nil
}

type baseVectorsStream struct{}

var _ grpc.ServerStream = baseVectorsStream{}

func (s baseVectorsStream) SetHeader(metadata.MD) error  { return nil }
func (s baseVectorsStream) SendHeader(metadata.MD) error { return nil }
func (s baseVectorsStream) SetTrailer(metadata.MD)       {}
func (s baseVectorsStream) Context() context.Context     { return context.Background() }
func (s baseVectorsStream) SendMsg(interface{}) error    { return nil }
func (s baseVectorsStream) RecvMsg(interface{}) error    { return nil }

func assertIndexContainsExactlyIDs(t *testing.T, im *indexmanager.IndexManager, indexName string, ids []uint32) {
	t.Helper()
	index, indexExists := im.GetIndex(indexName)
	assert.True(t, indexExists)

	results := index.SearchKNN(sampleVectors[0], len(ids))
	assert.Len(t, results, len(ids))

	actualIDs := make([]uint32, len(results))
	for i, r := range results {
		actualIDs[i] = r.ID
	}

	for _, id := range ids {
		assert.Contains(t, actualIDs, id)
	}
}

func createTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "server_test")
	assert.NoError(t, err)
	return dir
}

func deleteDir(t *testing.T, dir string) {
	t.Helper()
	err := os.RemoveAll(dir)
	assert.NoError(t, err)
}
