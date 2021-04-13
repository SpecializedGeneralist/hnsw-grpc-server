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

package pkg

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/grpcapi"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/grpcutils"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/internal/hnswgo"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"
)

var _ grpcapi.ServerServer = &Server{}

type Server struct {
	dataPath string
	indices  Indices
	grpcapi.UnimplementedServerServer
}

func NewServer(storage string, indices Indices) *Server {
	return &Server{
		dataPath: storage,
		indices:  indices,
	}
}

// StartServer is used to start the gRPC server.
func (s *Server) StartServer(address, tlsCert, tlsKey string, tlsEnabled bool) {
	grpcServer := grpcutils.NewGRPCServer(grpcutils.GRPCServerConfig{
		TLSEnabled: tlsEnabled,
		TLSCert:    tlsCert,
		TLSKey:     tlsKey,
	})
	grpcapi.RegisterServerServer(grpcServer, s)
	grpcutils.RunGRPCServer(address, grpcServer)
}

var spaceTypeMap = map[grpcapi.CreateIndexRequest_SpaceType]string{
	grpcapi.CreateIndexRequest_L2:     "l2",
	grpcapi.CreateIndexRequest_IP:     "ip",
	grpcapi.CreateIndexRequest_COSINE: "cosine",
}

// CreateIndex makes a new index.
func (s *Server) CreateIndex(_ context.Context, request *grpcapi.CreateIndexRequest) (*emptypb.Empty, error) {
	log.Debug().Msg("Received request for index creation.")
	indexName := request.IndexName
	if !isValidIndexName(indexName) {
		return nil, fmt.Errorf("invalid index name [%s]", indexName)
	}
	if _, exists := s.indices[indexName]; exists {
		return nil, fmt.Errorf("index [%s] already exists", indexName)
	}
	spaceType, ok := spaceTypeMap[request.SpaceType]
	if !ok {
		return nil, fmt.Errorf("invalid space type [%d]", request.SpaceType)
	}
	h := hnswgo.New(
		int(request.Dim),
		int(request.M),
		int(request.EfConstruction),
		int(request.Seed),
		uint32(request.MaxElements),
		spaceType,
		request.AutoId,
	)
	s.indices[indexName] = h

	return &emptypb.Empty{}, nil
}

// DeleteIndex removes an index.
func (s *Server) DeleteIndex(_ context.Context, request *grpcapi.DeleteIndexRequest) (*emptypb.Empty, error) {
	log.Debug().Msg("Received request for index deletion.")
	indexName := request.IndexName
	if !isValidIndexName(indexName) {
		return nil, fmt.Errorf("invalid index name [%s]", indexName)
	}
	_, exists := s.indices[indexName]
	if !exists {
		return &emptypb.Empty{}, nil
	}
	delete(s.indices, indexName)
	err := s.removeIndexFromStorage(indexName)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) removeIndexFromStorage(indexName string) error {
	files, err := filepath.Glob(path.Join(s.dataPath, fmt.Sprintf("%s_*", indexName)))
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}

// InsertVector inserts a new vector in the given index.
func (s *Server) InsertVector(_ context.Context, request *grpcapi.InsertVectorRequest) (*grpcapi.InsertVectorReply, error) {
	log.Debug().Msg("Received request for vector inserting.")
	start := time.Now()
	indexName := request.IndexName
	if !isValidIndexName(indexName) {
		return nil, fmt.Errorf("invalid index name [%s]", indexName)
	}
	index, exists := s.indices[indexName]
	if !exists {
		return nil, fmt.Errorf("index [%s] doesn't exists", indexName)
	}

	id, err := index.AddPointAutoID(request.Vector.Value)
	if err != nil {
		return nil, err
	}

	return &grpcapi.InsertVectorReply{
		Id:   fmt.Sprintf("%d", id),
		Took: time.Since(start).Milliseconds(),
	}, nil
}

// InsertVectorWithId inserts a new vector in the given index.
func (s *Server) InsertVectorWithId(_ context.Context, request *grpcapi.InsertVectorWithIdRequest) (*grpcapi.InsertVectorWithIdReply, error) {
	log.Debug().Msg("Received request for vector inserting.")
	start := time.Now()
	indexName := request.IndexName
	if !isValidIndexName(indexName) {
		return nil, fmt.Errorf("invalid index name [%s]", indexName)
	}
	index, exists := s.indices[indexName]
	if !exists {
		return nil, fmt.Errorf("index [%s] doesn't exists", indexName)
	}

	err := index.AddPoint(request.Vector.Value, uint32(request.Id))
	if err != nil {
		return nil, err
	}

	return &grpcapi.InsertVectorWithIdReply{
		Took: time.Since(start).Milliseconds(),
	}, nil
}

// InsertVectors inserts the new vectors in the given index. It flushes the index at each batch.
func (s *Server) InsertVectors(stream grpcapi.Server_InsertVectorsServer) error {
	log.Debug().Msg("Received request for vectors inserting.")
	start := time.Now()
	ids := make([]string, 0)
	var index *hnswgo.HNSW
	var indexName string
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			if index != nil {
				_ = s.removeIndexFromStorage(indexName) // important
				index.Save(path.Join(s.dataPath, makeIndexFilename(index, indexName)))
			}
			return stream.SendAndClose(&grpcapi.InsertVectorsReply{
				Ids:  ids,
				Took: time.Since(start).Milliseconds(),
			})
		}
		if err != nil {
			return err
		}
		curIndexName := request.IndexName
		if !isValidIndexName(curIndexName) {
			return fmt.Errorf("invalid index name [%s]", curIndexName)
		}
		curIndex, exists := s.indices[curIndexName]
		if !exists {
			return fmt.Errorf("index [%s] doesn't exists", curIndexName)
		}
		if indexName == "" {
			index = curIndex
			indexName = curIndexName
		}
		if indexName != curIndexName {
			return fmt.Errorf("the index must be constant during a batch insertion; expected [%s] found [%s]",
				indexName, curIndexName)
		}

		newID, err := index.AddPointAutoID(request.Vector.Value)
		if err != nil {
			return err
		}

		ids = append(ids, fmt.Sprintf("%d", newID))
	}
}

// InsertVectorsWithIds inserts the new vectors in the given index. It flushes the index at each batch.
func (s *Server) InsertVectorsWithIds(stream grpcapi.Server_InsertVectorsWithIdsServer) error {
	log.Debug().Msg("Received request for vectors inserting.")
	start := time.Now()
	var index *hnswgo.HNSW
	var indexName string
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			if index != nil {
				_ = s.removeIndexFromStorage(indexName) // important
				index.Save(path.Join(s.dataPath, makeIndexFilename(index, indexName)))
			}
			return stream.SendAndClose(&grpcapi.InsertVectorsWithIdsReply{
				Took: time.Since(start).Milliseconds(),
			})
		}
		if err != nil {
			return err
		}
		curIndexName := request.IndexName
		if !isValidIndexName(curIndexName) {
			return fmt.Errorf("invalid index name [%s]", curIndexName)
		}
		curIndex, exists := s.indices[curIndexName]
		if !exists {
			return fmt.Errorf("index [%s] doesn't exists", curIndexName)
		}
		if indexName == "" {
			index = curIndex
			indexName = curIndexName
		}
		if indexName != curIndexName {
			return fmt.Errorf("the index must be constant during a batch insertion; expected [%s] found [%s]",
				indexName, curIndexName)
		}

		err = index.AddPoint(request.Vector.Value, uint32(request.Id))
		if err != nil {
			return err
		}
	}
}

// SearchKNN returns the top k nearest neighbors to the query, searching on the given index.
func (s *Server) SearchKNN(_ context.Context, request *grpcapi.SearchRequest) (*grpcapi.SearchKNNReply, error) {
	log.Debug().Msg("Received request for searching.")
	start := time.Now()
	nameName := request.IndexName
	if !isValidIndexName(nameName) {
		return nil, fmt.Errorf("invalid index name [%s]", nameName)
	}
	index, exists := s.indices[nameName]
	if !exists {
		return nil, fmt.Errorf("index [%s] doesn't exists", nameName)
	}

	ids, dists := index.SearchKNN(request.Vector.Value, int(request.K))

	hits := make([]*grpcapi.Hit, len(ids))
	for i, id := range ids {
		hits[i] = &grpcapi.Hit{
			Id:       fmt.Sprintf("%d", id),
			Distance: dists[i],
		}
	}

	return &grpcapi.SearchKNNReply{
		Hits: hits,
		Took: time.Since(start).Milliseconds(),
	}, nil
}

// FlushIndex flushes the index to file.
func (s *Server) FlushIndex(_ context.Context, request *grpcapi.FlushRequest) (*emptypb.Empty, error) {
	log.Debug().Msg("Received request for index flushing.")
	indexName := request.IndexName
	if !isValidIndexName(indexName) {
		return nil, fmt.Errorf("invalid index name [%s]", indexName)
	}
	index, exists := s.indices[indexName]
	if !exists {
		return nil, fmt.Errorf("index [%s] doesn't exists", indexName)
	}

	_ = s.removeIndexFromStorage(indexName) // important
	index.Save(path.Join(s.dataPath, makeIndexFilename(index, indexName)))
	return &emptypb.Empty{}, nil
}

func makeIndexFilename(index *hnswgo.HNSW, indexName string) string {
	return fmt.Sprintf("%s_%s_%d_%d_%t",
		indexName,
		index.SpaceType,
		index.Dim,
		index.LastID,
		index.AutoID,
	)
}

// Indices returns the list of indices.
func (s *Server) Indices(_ context.Context, _ *emptypb.Empty) (*grpcapi.IndicesReply, error) {
	log.Debug().Msg("Received request for getting indices.")
	indices := make([]string, 0, len(s.indices))
	for key := range s.indices {
		indices = append(indices, key)
	}
	return &grpcapi.IndicesReply{
		Indices: indices,
	}, nil
}

// SetEf sets the `ef` parameter for the given index.
func (s *Server) SetEf(_ context.Context, request *grpcapi.SetEfRequest) (*emptypb.Empty, error) {
	log.Debug().Msg("Received request for `ef` setting.")
	indexName := request.IndexName
	if !isValidIndexName(indexName) {
		return nil, fmt.Errorf("invalid index name [%s]", indexName)
	}
	index, exists := s.indices[indexName]
	if !exists {
		return nil, fmt.Errorf("index [%s] doesn't exists", indexName)
	}
	index.SetEf(int(request.Value))
	return &emptypb.Empty{}, nil
}

var indexNameRegexp = regexp.MustCompile("^[a-z][a-z\\d_]*$")

func isValidIndexName(str string) bool {
	return len(str) <= 255 && indexNameRegexp.MatchString(str)
}
