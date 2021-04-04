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

package app

import (
	"fmt"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

const (
	programName = "hnsw-grpc-server"
)

// App contains everything needed to run the server.
type App struct {
	*cli.App
	address    string
	tlsCert    string
	tlsKey     string
	tlsEnabled bool
	debug      bool
	dataPath   string
}

// NewApp returns a new App objects.
func NewApp() *App {
	app := &App{
		App: cli.NewApp(),
	}
	app.Name = programName
	app.Usage = "A cli for the hnsw-grpc-server."
	app.Flags = flagsFor(app)
	app.Action = ActionFor(app)
	return app
}

func flagsFor(app *App) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "address",
			Value:       "0.0.0.0:19530",
			Usage:       "server binding address",
			Destination: &app.address,
		},
		&cli.StringFlag{
			Name:        "data",
			Value:       "/data",
			Usage:       "path to the indices folder",
			Destination: &app.dataPath,
		},
		&cli.BoolFlag{
			Name:        "tls",
			Value:       false,
			Usage:       "whether to use TLS",
			Destination: &app.tlsEnabled,
		},
		&cli.StringFlag{
			Name:        "tls-cert",
			Value:       "server.crt",
			Usage:       "TLS cert file",
			Destination: &app.tlsCert,
		},
		&cli.StringFlag{
			Name:        "tls-key",
			Value:       "server.key",
			Usage:       "TLS key file",
			Destination: &app.tlsKey,
		},
		&cli.BoolFlag{
			Name:        "debug",
			Value:       false,
			Usage:       "set the log level to debug",
			Destination: &app.debug,
		},
	}
}

func ActionFor(app *App) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		})

		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		if app.debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}

		if app.tlsEnabled {
			fmt.Printf("TLS Cert path is %s\n", app.tlsCert)
			fmt.Printf("TLS private key path is %s\n", app.tlsKey)
		}

		indices, err := pkg.Load(app.dataPath)
		if err != nil {
			return fmt.Errorf("error on loading indices from %s", app.dataPath)
		}

		server := pkg.NewServer(app.dataPath, indices)
		server.StartServer(app.address, app.tlsCert, app.tlsKey, app.tlsEnabled)
		return nil
	}
}
