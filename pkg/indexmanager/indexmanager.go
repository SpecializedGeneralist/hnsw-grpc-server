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

package indexmanager

import (
	"fmt"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/hnswgo"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/osutils"
	"github.com/rs/zerolog"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
)

var indexNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// IndexManager allows easy handling of multiple HNSW indices.
type IndexManager struct {
	path    string
	logger  zerolog.Logger
	indices map[string]*hnswgo.HNSW
	rwMx    sync.RWMutex
}

// New creates a new IndexManager.
func New(path string, logger zerolog.Logger) *IndexManager {
	return &IndexManager{
		path:    path,
		logger:  logger,
		indices: make(map[string]*hnswgo.HNSW),
		rwMx:    sync.RWMutex{},
	}
}

// LoadIndices load all HNSW indices stored in the configured path.
func (im *IndexManager) LoadIndices() error {
	im.rwMx.Lock()
	defer im.rwMx.Unlock()

	im.logger.Info().Msgf("loading all indices from dir %#v...", im.path)
	files, err := os.ReadDir(im.path)
	if err != nil {
		return fmt.Errorf("error reading content of indices dir %#v: %w", im.path, err)
	}

	for _, file := range files {
		name := file.Name()

		// Ignore files and hidden dirs
		if !file.IsDir() || strings.HasPrefix(name, ".") {
			continue
		}

		im.logger.Info().Msgf("loading index %#v...", name)
		err = im.loadIndex(name)
		if err != nil {
			return err
		}
	}

	im.logger.Info().Msg("all indices successfully loaded")
	return nil
}

// GetIndex returns the HNSW index (if it exists) and reports whether it is found.
func (im *IndexManager) GetIndex(name string) (*hnswgo.HNSW, bool) {
	im.rwMx.RLock()
	defer im.rwMx.RUnlock()

	index, found := im.indices[name]
	return index, found
}

// CreateIndex creates and persists a new index with the given name.
// If the name is not acceptable or an index with the same name already
// exists, an error is returned.
func (im *IndexManager) CreateIndex(name string, config hnswgo.Config) (*hnswgo.HNSW, error) {
	im.rwMx.Lock()
	defer im.rwMx.Unlock()

	if !isValidIndexName(name) {
		return nil, fmt.Errorf("invalid index name")
	}
	if _, ok := im.indices[name]; ok {
		return nil, fmt.Errorf("index with name %#v already exists", name)
	}

	dir := path.Join(im.path, name)
	dirExists, err := osutils.DirExists(dir)
	if err != nil {
		return nil, err
	}
	if dirExists {
		return nil, fmt.Errorf("index dir %#v already exists", dir)
	}

	index := hnswgo.New(dir, config)
	im.indices[name] = index

	err = index.Save()
	if err != nil {
		return nil, fmt.Errorf("error persisting new index %#v: %w", name, err)
	}

	return index, nil
}

// PersistIndex saves the current index to disk.
func (im *IndexManager) PersistIndex(name string) error {
	im.rwMx.RLock()
	defer im.rwMx.RUnlock()

	index, indexExists := im.indices[name]
	if !indexExists {
		return fmt.Errorf("index does not exist")
	}
	err := index.Save()
	if err != nil {
		return fmt.Errorf("error persisting index %#v: %w", name, err)
	}
	return nil
}

// DeleteIndex remove an index, also removing data from disk.
func (im *IndexManager) DeleteIndex(name string) error {
	im.rwMx.Lock()
	defer im.rwMx.Unlock()

	if _, ok := im.indices[name]; !ok {
		return fmt.Errorf("index does not exist")
	}

	filename := path.Join(im.path, name)
	dirExists, err := osutils.DirExists(filename)
	if err != nil {
		return err
	}
	if dirExists {
		err = os.RemoveAll(filename)
		if err != nil {
			return fmt.Errorf("error removing index dir %#v: %w", filename, err)
		}
	}

	delete(im.indices, name)
	return nil
}

// IndicesNames returns the names of all indices.
func (im *IndexManager) IndicesNames() []string {
	im.rwMx.RLock()
	defer im.rwMx.RUnlock()

	names := make([]string, 0, len(im.indices))
	for name := range im.indices {
		names = append(names, name)
	}
	return names
}

// Size returns the amount of currently loaded indices.
func (im *IndexManager) Size() int {
	im.rwMx.RLock()
	defer im.rwMx.RUnlock()

	return len(im.indices)
}

func (im *IndexManager) loadIndex(name string) error {
	if _, ok := im.indices[name]; ok {
		return fmt.Errorf("index %#v was already loaded", name)
	}

	h, err := hnswgo.Load(path.Join(im.path, name))
	if err != nil {
		return fmt.Errorf("error loading index %#v: %w", name, err)
	}
	im.indices[name] = h
	return nil
}

func isValidIndexName(name string) bool {
	return len(name) <= 255 && indexNameRegexp.MatchString(name)
}
