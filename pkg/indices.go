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
	"fmt"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/internal/hnswgo"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"strconv"
	"strings"
)

type Indices = map[string]*hnswgo.HNSW

func Load(dataPath string) (Indices, error) {
	indices := make(Indices)
	f, err := os.Open(dataPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	fis, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	for _, fi := range fis {
		log.Info().Msg(fmt.Sprintf("Loading index `%s`.", fi.Name()))
		fname := fi.Name()
		spl := strings.Split(fname, "_")
		indexName := spl[0]
		spaceType := spl[1]
		dim, err := strconv.Atoi(spl[2])
		if err != nil {
			return nil, fmt.Errorf("invalid filename %s", fname)
		}
		lastID, err := strconv.Atoi(spl[3])
		if err != nil {
			return nil, fmt.Errorf("invalid filename %s", fname)
		}
		autoID := spl[4]

		h := hnswgo.Load(path.Join(dataPath, fname), dim, spaceType, autoID)
		h.LastID = uint32(lastID)
		indices[indexName] = h
	}
	return indices, nil
}
