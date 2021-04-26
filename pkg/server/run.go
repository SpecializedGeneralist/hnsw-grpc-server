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
	"fmt"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/grpcapi"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"net"
	"runtime/debug"
)

// Run runs the server according to the configuration.
func (s *Server) Run() error {
	serverOptions, err := s.createServerOptions()
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(serverOptions...)
	grpcapi.RegisterServerServer(grpcServer, s)

	listener, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return fmt.Errorf("TCP listen error: %w", err)
	}

	s.logger.Info().Msgf("Serving on %s", s.config.Address)
	return grpcServer.Serve(listener)
}

func (s *Server) createServerOptions() ([]grpc.ServerOption, error) {
	options := []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpcrecovery.UnaryServerInterceptor(
				grpcrecovery.WithRecoveryHandler(func(p interface{}) error {
					s.logger.Error().Msgf("Panic! Stack trace:\n%s", string(debug.Stack()))
					return status.Errorf(codes.Internal, "panic: %v", p)
				}),
			),
		),
	}

	if s.config.TLSEnabled {
		s.logger.Info().Str("cert", s.config.TLSCert).Str("key", s.config.TLSKey).Msgf("TLS enabled")

		creds, err := credentials.NewServerTLSFromFile(s.config.TLSCert, s.config.TLSKey)
		if err != nil {
			return nil, fmt.Errorf("failed to read TLS certs: %w", err)
		}
		options = append(options, grpc.Creds(creds))
	}

	return options, nil
}
