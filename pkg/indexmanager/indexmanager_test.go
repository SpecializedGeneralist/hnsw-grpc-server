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

package indexmanager_test

import (
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/hnswgo"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/indexmanager"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

var sampleConfig = hnswgo.Config{
	SpaceType:      hnswgo.CosineSpace,
	Dim:            5,
	MaxElements:    10,
	M:              10,
	EfConstruction: 200,
	RandSeed:       100,
	AutoIDEnabled:  true,
}

func TestIndexManager_LoadIndices(t *testing.T) {
	t.Parallel()

	t.Run("missing directory", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		im := indexmanager.New(path.Join(dir, "foo"), zerolog.Nop())
		err := im.LoadIndices()
		assert.Error(t, err)
	})

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		im := indexmanager.New(dir, zerolog.Nop())
		err := im.LoadIndices()
		assert.NoError(t, err)
		assert.Equal(t, 0, im.Size())
	})

	t.Run("it ignores files and hidden directories", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		file, err := os.Create(path.Join(dir, "file"))
		assert.NoError(t, err)
		err = file.Close()
		assert.NoError(t, err)

		err = os.Mkdir(path.Join(dir, ".hidden-dir"), 0777)
		assert.NoError(t, err)

		im := indexmanager.New(dir, zerolog.Nop())
		err = im.LoadIndices()
		assert.NoError(t, err)
		assert.Equal(t, 0, im.Size())
	})

	t.Run("indices with errors", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		// An empty directory is surely not a valid index
		err := os.Mkdir(path.Join(dir, "not-an-index"), 0777)
		assert.NoError(t, err)

		im := indexmanager.New(dir, zerolog.Nop())
		err = im.LoadIndices()
		assert.Error(t, err)
	})

	t.Run("existing indices", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		createAndSaveSampleIndex(t, path.Join(dir, "foo"))
		createAndSaveSampleIndex(t, path.Join(dir, "bar"))

		im := indexmanager.New(dir, zerolog.Nop())
		err := im.LoadIndices()
		assert.NoError(t, err)
		assert.Equal(t, 2, im.Size())

		names := im.IndicesNames()
		assert.Len(t, names, 2)
		assert.Contains(t, names, "foo")
		assert.Contains(t, names, "bar")

		index, found := im.GetIndex("foo")
		assert.NotNil(t, index)
		assert.True(t, found)

		index, found = im.GetIndex("bar")
		assert.NotNil(t, index)
		assert.True(t, found)

		index, found = im.GetIndex("baz")
		assert.Nil(t, index)
		assert.False(t, found)
	})
}

func TestIndexManager_CreateIndex(t *testing.T) {
	t.Parallel()

	t.Run("creating valid index", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())

		index, err := im.CreateIndex("foo", sampleConfig)
		assert.NoError(t, err)
		assert.NotNil(t, index)

		index, err = im.CreateIndex("bar", sampleConfig)
		assert.NoError(t, err)
		assert.NotNil(t, index)

		names := im.IndicesNames()
		assert.Len(t, names, 2)
		assert.Contains(t, names, "foo")
		assert.Contains(t, names, "bar")

		index, found := im.GetIndex("foo")
		assert.NotNil(t, index)
		assert.True(t, found)

		index, found = im.GetIndex("bar")
		assert.NotNil(t, index)
		assert.True(t, found)

		index, found = im.GetIndex("baz")
		assert.Nil(t, index)
		assert.False(t, found)
	})

	t.Run("invalid name", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())
		index, err := im.CreateIndex("foo!?", sampleConfig)
		assert.Error(t, err)
		assert.Nil(t, index)
	})

	t.Run("index already exists", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())

		index, err := im.CreateIndex("foo", sampleConfig)
		assert.NoError(t, err)
		assert.NotNil(t, index)

		index, err = im.CreateIndex("foo", sampleConfig)
		assert.Error(t, err)
		assert.Nil(t, index)
	})

	t.Run("index dir already exists", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())

		err := os.Mkdir(path.Join(dir, "foo"), 0777)
		assert.NoError(t, err)

		index, err := im.CreateIndex("foo", sampleConfig)
		assert.Error(t, err)
		assert.Nil(t, index)
	})

	t.Run("index dir check error", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())

		file, err := os.Create(path.Join(dir, "foo"))
		assert.NoError(t, err)
		err = file.Close()
		assert.NoError(t, err)

		index, err := im.CreateIndex("foo", sampleConfig)
		assert.Error(t, err)
		assert.Nil(t, index)
	})
}

func TestIndexManager_PersistIndex(t *testing.T) {
	t.Parallel()

	t.Run("successful persistence", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		{
			im := indexmanager.New(dir, zerolog.Nop())

			foo, err := im.CreateIndex("foo", sampleConfig)
			assert.NoError(t, err)
			_, err = foo.AddPointAutoID(sampleVectors[0])
			assert.NoError(t, err)

			bar, err := im.CreateIndex("bar", sampleConfig)
			assert.NoError(t, err)
			_, err = bar.AddPointAutoID(sampleVectors[1])
			assert.NoError(t, err)

			err = im.PersistIndex("foo")
			assert.NoError(t, err)

			err = im.PersistIndex("bar")
			assert.NoError(t, err)
		}
		{
			im := indexmanager.New(dir, zerolog.Nop())
			err := im.LoadIndices()
			assert.NoError(t, err)
			assert.Equal(t, 2, im.Size())

			names := im.IndicesNames()
			assert.Len(t, names, 2)
			assert.Contains(t, names, "foo")
			assert.Contains(t, names, "bar")

			index, found := im.GetIndex("foo")
			assert.NotNil(t, index)
			assert.True(t, found)

			index, found = im.GetIndex("bar")
			assert.NotNil(t, index)
			assert.True(t, found)
		}
	})

	t.Run("index does not exist", func(t *testing.T) {
		t.Parallel()
		im := indexmanager.New(os.TempDir(), zerolog.Nop())

		err := im.PersistIndex("foo")
		assert.Error(t, err)
	})

	t.Run("saving error", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		subDir := path.Join(dir, "sub")
		err := os.Mkdir(subDir, 0777)
		assert.NoError(t, err)

		im := indexmanager.New(subDir, zerolog.Nop())

		_, err = im.CreateIndex("foo", sampleConfig)
		assert.NoError(t, err)

		deleteDir(t, subDir)

		err = im.PersistIndex("foo")
		assert.Error(t, err)
	})
}

func TestIndexManager_DeleteIndex(t *testing.T) {
	t.Parallel()

	t.Run("deleting non persisted index", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())

		_, err := im.CreateIndex("bar", sampleConfig)
		assert.NoError(t, err)

		names := im.IndicesNames()
		assert.Equal(t, []string{"bar"}, names)

		index, found := im.GetIndex("bar")
		assert.NotNil(t, index)
		assert.True(t, found)

		err = im.DeleteIndex("bar")
		assert.NoError(t, err)

		names = im.IndicesNames()
		assert.Empty(t, names)

		index, found = im.GetIndex("bar")
		assert.Nil(t, index)
		assert.False(t, found)
	})

	t.Run("deleting persisted index", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		im := indexmanager.New(dir, zerolog.Nop())

		_, err := im.CreateIndex("foo", sampleConfig)
		assert.NoError(t, err)
		err = im.PersistIndex("foo")
		assert.NoError(t, err)

		_, err = im.CreateIndex("bar", sampleConfig)
		assert.NoError(t, err)
		err = im.PersistIndex("bar")
		assert.NoError(t, err)

		names := im.IndicesNames()
		assert.Len(t, names, 2)
		assert.Contains(t, names, "foo")
		assert.Contains(t, names, "bar")

		index, found := im.GetIndex("foo")
		assert.NotNil(t, index)
		assert.True(t, found)

		index, found = im.GetIndex("bar")
		assert.NotNil(t, index)
		assert.True(t, found)

		assert.DirExists(t, path.Join(dir, "foo"))
		assert.DirExists(t, path.Join(dir, "bar"))

		err = im.DeleteIndex("foo")
		assert.NoError(t, err)

		names = im.IndicesNames()
		assert.Equal(t, []string{"bar"}, names)

		index, found = im.GetIndex("foo")
		assert.Nil(t, index)
		assert.False(t, found)

		index, found = im.GetIndex("bar")
		assert.NotNil(t, index)
		assert.True(t, found)

		assert.NoDirExists(t, path.Join(dir, "foo"))
		assert.DirExists(t, path.Join(dir, "bar"))
	})

	t.Run("index does not exist", func(t *testing.T) {
		t.Parallel()
		im := indexmanager.New(os.TempDir(), zerolog.Nop())

		err := im.DeleteIndex("foo")
		assert.Error(t, err)
	})
}

var sampleVectors = [][]float32{
	{0.1, 0.2, 0.3, 0.4, 0.5},
	{0.9, 0.8, 0.7, 0.6, 0.5},
}

func createAndSaveSampleIndex(t *testing.T, dir string) {
	hnsw := hnswgo.New(dir, sampleConfig)
	for _, vector := range sampleVectors {
		_, err := hnsw.AddPointAutoID(vector)
		assert.NoError(t, err)
	}
	err := hnsw.Save()
	assert.NoError(t, err)

	assert.FileExists(t, path.Join(dir, "state"))
	assert.FileExists(t, path.Join(dir, "index"))
}

func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "indexmanager_test")
	assert.NoError(t, err)
	return dir
}

func deleteDir(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	assert.NoError(t, err)
}
