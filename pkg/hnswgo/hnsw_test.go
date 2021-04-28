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

package hnswgo_test

import (
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/hnswgo"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

var sampleVectors = [][]float32{
	{0.1, 0.2, 0.3, 0.4, 0.5},
	{0.9, 0.8, 0.7, 0.6, 0.5},
}

func TestSpaceTypeFromString(t *testing.T) {
	t.Parallel()

	t.Run("IPSpace", func(t *testing.T) {
		t.Parallel()
		val, err := hnswgo.SpaceTypeFromString("ip")
		assert.NoError(t, err)
		assert.Equal(t, hnswgo.IPSpace, val)
	})

	t.Run("CosineSpace", func(t *testing.T) {
		t.Parallel()
		val, err := hnswgo.SpaceTypeFromString("cosine")
		assert.NoError(t, err)
		assert.Equal(t, hnswgo.CosineSpace, val)
	})

	t.Run("L2Space", func(t *testing.T) {
		t.Parallel()
		val, err := hnswgo.SpaceTypeFromString("l2")
		assert.NoError(t, err)
		assert.Equal(t, hnswgo.L2Space, val)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		_, err := hnswgo.SpaceTypeFromString("foo")
		assert.Error(t, err)
	})
}

func TestHNSW_IPSpace(t *testing.T) {
	t.Parallel()

	hnsw := hnswgo.New(makeConfig(hnswgo.IPSpace, false))

	for i, vector := range sampleVectors {
		err := hnsw.AddPoint(vector, uint32(i))
		assert.NoError(t, err)
	}

	results := hnsw.SearchKNN(sampleVectors[0], 2)
	assert.Len(t, results, 2)
}

func TestHNSW_L2Space(t *testing.T) {
	t.Parallel()

	hnsw := hnswgo.New(makeConfig(hnswgo.L2Space, false))

	for i, vector := range sampleVectors {
		err := hnsw.AddPoint(vector, uint32(i))
		assert.NoError(t, err)
	}

	results := hnsw.SearchKNN(sampleVectors[0], 2)
	assert.Len(t, results, 2)
}

func TestHNSW_CosineSpace(t *testing.T) {
	t.Parallel()

	hnsw := hnswgo.New(makeConfig(hnswgo.CosineSpace, false))

	for i, vector := range sampleVectors {
		err := hnsw.AddPoint(vector, uint32(i))
		assert.NoError(t, err)
	}

	results := hnsw.SearchKNN(sampleVectors[0], 2)
	assert.Len(t, results, 2)
	assert.Equal(t, uint32(0), results[0].ID)
	assert.InDelta(t, 0.0, results[0].Distance, 1e-6)
	assert.Equal(t, uint32(1), results[1].ID)
	assert.Greater(t, results[1].Distance, results[0].Distance)

	results = hnsw.SearchKNN(sampleVectors[1], 2)
	assert.Len(t, results, 2)
	assert.Equal(t, uint32(1), results[0].ID)
	assert.InDelta(t, 0.0, results[0].Distance, 1e-6)
	assert.Equal(t, uint32(0), results[1].ID)
	assert.Greater(t, results[1].Distance, results[0].Distance)
}

func TestHNSW_AutoIDDisabled(t *testing.T) {
	t.Parallel()

	hnsw := hnswgo.New(makeConfig(hnswgo.CosineSpace, false))

	err := hnsw.AddPoint(sampleVectors[0], uint32(42))
	assert.NoError(t, err)

	_, err = hnsw.AddPointAutoID(sampleVectors[1])
	assert.Error(t, err)

	results := hnsw.SearchKNN(sampleVectors[0], 1)
	assert.Len(t, results, 1)
	assert.Equal(t, uint32(42), results[0].ID)
	assert.InDelta(t, 0.0, results[0].Distance, 1e-6)
}

func TestHNSW_AutoIDEnabled(t *testing.T) {
	t.Parallel()

	hnsw := hnswgo.New(makeConfig(hnswgo.CosineSpace, true))

	for i, vector := range sampleVectors {
		id, err := hnsw.AddPointAutoID(vector)
		assert.NoError(t, err)
		assert.Equal(t, i+1, int(id))
	}

	err := hnsw.AddPoint(sampleVectors[0], uint32(42))
	assert.Error(t, err)

	results := hnsw.SearchKNN(sampleVectors[0], 2)
	assert.Equal(t, uint32(1), results[0].ID)
	assert.Equal(t, uint32(2), results[1].ID)
}

func TestHNSW_MarkDelete(t *testing.T) {
	t.Parallel()

	hnsw := hnswgo.New(makeConfig(hnswgo.CosineSpace, false))

	for i, vector := range sampleVectors {
		err := hnsw.AddPoint(vector, uint32(i))
		assert.NoError(t, err)
	}

	results := hnsw.SearchKNN(sampleVectors[0], 1)
	assert.Equal(t, uint32(0), results[0].ID)

	hnsw.MarkDelete(0)

	results = hnsw.SearchKNN(sampleVectors[0], 1)
	assert.Equal(t, uint32(1), results[0].ID)
}

func TestHNSW_SaveAndLoad(t *testing.T) {
	t.Parallel()

	dir := createTempDir(t)
	defer deleteDir(t, dir)

	var originalResults []hnswgo.KNNResult

	{
		hnsw := hnswgo.New(makeConfig(hnswgo.CosineSpace, false))

		for i, vector := range sampleVectors {
			err := hnsw.AddPoint(vector, uint32(i))
			assert.NoError(t, err)
		}

		originalResults = hnsw.SearchKNN(sampleVectors[0], 2)
		assert.Len(t, originalResults, 2)

		err := hnsw.Save(dir)
		assert.NoError(t, err)
	}

	hnsw, err := hnswgo.Load(dir)
	assert.NoError(t, err)

	newResults := hnsw.SearchKNN(sampleVectors[0], 2)
	assert.Equal(t, originalResults, newResults)
}

func TestHNSW_Save(t *testing.T) {
	t.Parallel()

	t.Run("path does not exist", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		hnsw := hnswgo.New(makeConfig(hnswgo.CosineSpace, true))
		err := hnsw.Save(path.Join(dir, "foo", "bar"))
		assert.Error(t, err)
	})
}

func TestHNSW_LoadingErrors(t *testing.T) {
	t.Parallel()

	t.Run("state.tmp exists", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		createAndSaveSampleIndex(t, dir)
		err := os.Rename(path.Join(dir, "state"), path.Join(dir, "state.tmp"))
		assert.NoError(t, err)

		hnsw, err := hnswgo.Load(dir)
		assert.Nil(t, hnsw)
		assert.Error(t, err)
	})

	t.Run("state does not exist", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		createAndSaveSampleIndex(t, dir)
		err := os.Remove(path.Join(dir, "state"))
		assert.NoError(t, err)

		hnsw, err := hnswgo.Load(dir)
		assert.Nil(t, hnsw)
		assert.Error(t, err)
	})

	t.Run("index.tmp exists", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		createAndSaveSampleIndex(t, dir)
		err := os.Rename(path.Join(dir, "index"), path.Join(dir, "index.tmp"))
		assert.NoError(t, err)

		hnsw, err := hnswgo.Load(dir)
		assert.Nil(t, hnsw)
		assert.Error(t, err)
	})

	t.Run("index does not exist", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		createAndSaveSampleIndex(t, dir)
		err := os.Remove(path.Join(dir, "index"))
		assert.NoError(t, err)

		hnsw, err := hnswgo.Load(dir)
		assert.Nil(t, hnsw)
		assert.Error(t, err)
	})

	t.Run("the path does not exist", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		hnsw, err := hnswgo.Load(path.Join(dir, "foo"))
		assert.Nil(t, hnsw)
		assert.Error(t, err)
	})
}

func createAndSaveSampleIndex(t *testing.T, dir string) {
	hnsw := hnswgo.New(makeConfig(hnswgo.CosineSpace, true))
	for i, vector := range sampleVectors {
		id, err := hnsw.AddPointAutoID(vector)
		assert.NoError(t, err)
		assert.Equal(t, i+1, int(id))
	}
	err := hnsw.Save(dir)
	assert.NoError(t, err)

	assert.FileExists(t, path.Join(dir, "state"))
	assert.FileExists(t, path.Join(dir, "index"))
}

func makeConfig(spaceType hnswgo.SpaceType, autoIDEnabled bool) hnswgo.Config {
	return hnswgo.Config{
		SpaceType:      spaceType,
		Dim:            5,
		MaxElements:    10,
		M:              10,
		EfConstruction: 200,
		RandSeed:       100,
		AutoIDEnabled:  autoIDEnabled,
	}
}

func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "hnsw_test")
	assert.NoError(t, err)
	return dir
}

func deleteDir(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	assert.NoError(t, err)
}
