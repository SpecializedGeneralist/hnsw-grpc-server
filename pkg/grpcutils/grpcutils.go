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

package grpcutils

import (
	"fmt"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"net"
	"runtime/debug"
)

// GRPCServerConfig provides server configuration parameters for creating
// a GRPC server (see NewGRPCServer).
type GRPCServerConfig struct {
	TLSEnabled bool
	TLSCert    string
	TLSKey     string
}

// NewGRPCServer returns grpc.Server objects, optionally configured for TLS.
func NewGRPCServer(config GRPCServerConfig) *grpc.Server {
	serverOptions := createServerOptions(config)
	return grpc.NewServer(serverOptions...)
}

func createServerOptions(config GRPCServerConfig) []grpc.ServerOption {
	options := []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpcrecovery.UnaryServerInterceptor(
				grpcrecovery.WithRecoveryHandler(func(p interface{}) error {
					log.Error().Err(fmt.Errorf("%s", string(debug.Stack()))).Msg("StackTrace:")
					return status.Errorf(codes.Internal, "panic: %v", p)
				}),
			),
		),
	}

	if config.TLSEnabled {
		creds, err := credentials.NewServerTLSFromFile(config.TLSCert, config.TLSKey)
		if err != nil {
			log.Fatal().Msg(fmt.Sprintf("failed to read TLS certs: %v\n", err))
		}
		options = append(options, grpc.Creds(creds))
	}

	return options
}

// RunGRPCServer listens on the given address and serves the given *grpc.Server,
// and blocks until done.
func RunGRPCServer(grpcAddress string, grpcServer *grpc.Server) {
	log.Info().Msgf("Starting: gRPC Listener [%s]", grpcAddress)
	listener := newListenerForGRPC(grpcAddress)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve")
	}
}

func newListenerForGRPC(grpcAddress string) net.Listener {
	result, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("failed to listen: %v", err))
	}
	return result
}
