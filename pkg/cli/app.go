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

package cli

import (
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/indexmanager"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/server"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

// App contains everything needed to run the CLI server application.
type App struct {
	*cli.App
	serverConfig server.Config
	debug        bool
	dataPath     string
}

// NewApp returns a new App object.
func NewApp() *App {
	app := &App{
		App: &cli.App{
			HelpName:  "hnsw-grpc-server",
			Usage:     "HNSW gRPC server",
			Reader:    os.Stdin,
			Writer:    os.Stdout,
			ErrWriter: os.Stderr,
		},
	}
	app.Flags = app.cliFlags()
	app.Action = app.runAction
	return app
}

func (app *App) cliFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "address",
			Value:       "0.0.0.0:19530",
			Usage:       "server binding address and port",
			Destination: &app.serverConfig.Address,
		},
		&cli.BoolFlag{
			Name:        "tls",
			Value:       false,
			Usage:       "whether to use TLS",
			Destination: &app.serverConfig.TLSEnabled,
		},
		&cli.StringFlag{
			Name:        "tls-cert",
			Value:       "server.crt",
			Usage:       "TLS cert file",
			Destination: &app.serverConfig.TLSCert,
		},
		&cli.StringFlag{
			Name:        "tls-key",
			Value:       "server.key",
			Usage:       "TLS key file",
			Destination: &app.serverConfig.TLSKey,
		},
		&cli.StringFlag{
			Name:        "data",
			Value:       "./hnsw-grpc-server-data",
			Usage:       "path to the indices folder",
			Destination: &app.dataPath,
		},
		&cli.BoolFlag{
			Name:        "debug",
			Value:       false,
			Usage:       "set the log level to debug",
			Destination: &app.debug,
		},
	}
}

func (app *App) runAction(*cli.Context) (err error) {
	logger := app.newLogger()

	defer func() {
		if err != nil {
			logger.Err(err).Send()
		}
	}()

	indexManager := indexmanager.New(app.dataPath, logger)
	err = indexManager.LoadIndices()
	if err != nil {
		return err
	}

	srv := server.New(app.serverConfig, indexManager, logger)
	return srv.Run()
}

func (app *App) newLogger() zerolog.Logger {
	level := zerolog.InfoLevel
	if app.debug {
		level = zerolog.DebugLevel
	}

	w := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}
	return zerolog.New(w).With().Timestamp().Logger().Level(level)
}
