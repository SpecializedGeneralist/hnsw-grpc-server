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

package server

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/grpcapi"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/hnswgo"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/indexmanager"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"strings"
	"time"
)

type Server struct {
	grpcapi.UnimplementedServerServer
	config       Config
	indexManager *indexmanager.IndexManager
	logger       zerolog.Logger
}

var _ grpcapi.ServerServer = &Server{}

func New(config Config, indexManager *indexmanager.IndexManager, logger zerolog.Logger) *Server {
	return &Server{
		config:       config,
		indexManager: indexManager,
		logger:       logger,
	}
}

var spaceTypeMap = map[grpcapi.CreateIndexRequest_SpaceType]hnswgo.SpaceType{
	grpcapi.CreateIndexRequest_L2:     hnswgo.L2Space,
	grpcapi.CreateIndexRequest_IP:     hnswgo.IPSpace,
	grpcapi.CreateIndexRequest_COSINE: hnswgo.CosineSpace,
}

// CreateIndex makes a new index.
func (s *Server) CreateIndex(_ context.Context, req *grpcapi.CreateIndexRequest) (*emptypb.Empty, error) {
	s.logger.Debug().Interface("req", req).Msg("Server.CreateIndex")

	spaceType, ok := spaceTypeMap[req.GetSpaceType()]
	if !ok {
		return nil, fmt.Errorf("invalid space type [%v]", req.GetSpaceType())
	}

	_, err := s.indexManager.CreateIndex(
		req.GetIndexName(),
		hnswgo.Config{
			SpaceType:      spaceType,
			Dim:            int(req.GetDim()),
			MaxElements:    int(req.GetMaxElements()),
			M:              int(req.GetM()),
			EfConstruction: int(req.GetEfConstruction()),
			RandSeed:       int(req.GetSeed()),
			AutoIDEnabled:  req.GetAutoId(),
		},
	)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// DeleteIndex removes an index.
func (s *Server) DeleteIndex(_ context.Context, req *grpcapi.DeleteIndexRequest) (*emptypb.Empty, error) {
	s.logger.Debug().Interface("req", req).Msg("Server.DeleteIndex")

	err := s.indexManager.DeleteIndex(req.GetIndexName())
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// InsertVector inserts a new vector in the given index.
func (s *Server) InsertVector(_ context.Context, req *grpcapi.InsertVectorRequest) (*grpcapi.InsertVectorReply, error) {
	s.logger.Debug().Interface("req", req).Msg("Server.InsertVector")

	startTime := time.Now()

	index, indexExists := s.indexManager.GetIndex(req.GetIndexName())
	if !indexExists {
		return nil, fmt.Errorf("index not found")
	}

	id, err := index.AddPointAutoID(req.GetVector().GetValue())
	if err != nil {
		return nil, err
	}
	return &grpcapi.InsertVectorReply{
		Id:   fmt.Sprintf("%d", id),
		Took: time.Since(startTime).Milliseconds(),
	}, nil
}

// InsertVectorWithId inserts a new vector in the given index.
func (s *Server) InsertVectorWithId(_ context.Context, req *grpcapi.InsertVectorWithIdRequest) (*grpcapi.InsertVectorWithIdReply, error) {
	s.logger.Debug().Interface("req", req).Msg("Server.InsertVectorWithId")

	startTime := time.Now()

	index, indexExists := s.indexManager.GetIndex(req.GetIndexName())
	if !indexExists {
		return nil, fmt.Errorf("index not found")
	}

	err := index.AddPoint(req.GetVector().GetValue(), uint32(req.GetId()))
	if err != nil {
		return nil, err
	}
	return &grpcapi.InsertVectorWithIdReply{
		Took: time.Since(startTime).Milliseconds(),
	}, nil
}

// InsertVectors inserts the new vectors in the given index. It flushes the index at each batch.
func (s *Server) InsertVectors(stream grpcapi.Server_InsertVectorsServer) error {
	s.logger.Debug().Msg("Server.InsertVectors")

	startTime := time.Now()

	ids := make([]string, 0)
	indicesNames := make(map[string]struct{})

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		indexName := req.GetIndexName()
		index, indexExists := s.indexManager.GetIndex(indexName)
		if !indexExists {
			return fmt.Errorf("index %#v not found", indexName)
		}

		newID, err := index.AddPointAutoID(req.GetVector().GetValue())
		if err != nil {
			return err
		}

		ids = append(ids, fmt.Sprintf("%d", newID))
		indicesNames[indexName] = struct{}{}
	}

	errors := make([]string, 0)
	for name := range indicesNames {
		err := s.indexManager.PersistIndex(name)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "\n"))
	}

	return stream.SendAndClose(&grpcapi.InsertVectorsReply{
		Ids:  ids,
		Took: time.Since(startTime).Milliseconds(),
	})
}

// InsertVectorsWithIds inserts the new vectors in the given index. It flushes the index at each batch.
func (s *Server) InsertVectorsWithIds(stream grpcapi.Server_InsertVectorsWithIdsServer) error {
	s.logger.Debug().Msg("Server.InsertVectorsWithIds")

	startTime := time.Now()

	indicesNames := make(map[string]struct{})

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		indexName := req.GetIndexName()
		index, indexExists := s.indexManager.GetIndex(indexName)
		if !indexExists {
			return fmt.Errorf("index %#v not found", indexName)
		}

		err = index.AddPoint(req.GetVector().GetValue(), uint32(req.GetId()))
		if err != nil {
			return err
		}
		indicesNames[indexName] = struct{}{}
	}

	errors := make([]string, 0)
	for name := range indicesNames {
		err := s.indexManager.PersistIndex(name)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "\n"))
	}

	return stream.SendAndClose(&grpcapi.InsertVectorsWithIdsReply{
		Took: time.Since(startTime).Milliseconds(),
	})
}

// SearchKNN returns the top k nearest neighbors to the query, searching on the given index.
func (s *Server) SearchKNN(_ context.Context, req *grpcapi.SearchRequest) (*grpcapi.SearchKNNReply, error) {
	s.logger.Debug().Interface("req", req).Msg("Server.SearchKNN")

	startTime := time.Now()

	index, indexExists := s.indexManager.GetIndex(req.GetIndexName())
	if !indexExists {
		return nil, fmt.Errorf("index not found")
	}

	results := index.SearchKNN(req.GetVector().GetValue(), int(req.GetK()))

	hits := make([]*grpcapi.Hit, len(results))
	for i, result := range results {
		hits[i] = &grpcapi.Hit{
			Id:       fmt.Sprintf("%d", result.ID),
			Distance: result.Distance,
		}
	}

	return &grpcapi.SearchKNNReply{
		Hits: hits,
		Took: time.Since(startTime).Milliseconds(),
	}, nil
}

// FlushIndex flushes the index to file.
func (s *Server) FlushIndex(_ context.Context, req *grpcapi.FlushRequest) (*emptypb.Empty, error) {
	s.logger.Debug().Interface("req", req).Msg("Server.FlushIndex")

	err := s.indexManager.PersistIndex(req.GetIndexName())
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// Indices returns the list of indices.
func (s *Server) Indices(context.Context, *emptypb.Empty) (*grpcapi.IndicesReply, error) {
	log.Debug().Msg("Received req for getting indices.")

	return &grpcapi.IndicesReply{
		Indices: s.indexManager.IndicesNames(),
	}, nil
}

// SetEf sets the `ef` parameter for the given index.
func (s *Server) SetEf(_ context.Context, req *grpcapi.SetEfRequest) (*emptypb.Empty, error) {
	log.Debug().Msg("Received req for `ef` setting.")

	index, indexExists := s.indexManager.GetIndex(req.GetIndexName())
	if !indexExists {
		return nil, fmt.Errorf("index not found")
	}
	index.SetEf(int(req.GetValue()))
	return &emptypb.Empty{}, nil
}
