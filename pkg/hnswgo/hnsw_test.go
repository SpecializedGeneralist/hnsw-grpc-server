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
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	dir := createTempDir(t)
	defer deleteDir(t, dir)

	hnsw := hnswgo.New(dir, makeConfig(hnswgo.IPSpace, false), zerolog.Nop())

	for i, vector := range sampleVectors {
		assert.NoError(t, hnsw.AddPoint(vector, uint32(i)))
	}

	results := hnsw.SearchKNN(sampleVectors[0], 2)
	assert.Len(t, results, 2)
}

func TestHNSW_L2Space(t *testing.T) {
	t.Parallel()
	dir := createTempDir(t)
	defer deleteDir(t, dir)

	hnsw := hnswgo.New(dir, makeConfig(hnswgo.L2Space, false), zerolog.Nop())

	for i, vector := range sampleVectors {
		assert.NoError(t, hnsw.AddPoint(vector, uint32(i)))
	}

	results := hnsw.SearchKNN(sampleVectors[0], 2)
	assert.Len(t, results, 2)
}

func TestHNSW_CosineSpace(t *testing.T) {
	t.Parallel()
	dir := createTempDir(t)
	defer deleteDir(t, dir)

	hnsw := hnswgo.New(dir, makeConfig(hnswgo.CosineSpace, false), zerolog.Nop())

	for i, vector := range sampleVectors {
		assert.NoError(t, hnsw.AddPoint(vector, uint32(i)))
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
	dir := createTempDir(t)
	defer deleteDir(t, dir)

	hnsw := hnswgo.New(dir, makeConfig(hnswgo.CosineSpace, false), zerolog.Nop())

	assert.NoError(t, hnsw.AddPoint(sampleVectors[0], uint32(42)))

	_, err := hnsw.AddPointAutoID(sampleVectors[1])
	assert.Error(t, err)

	results := hnsw.SearchKNN(sampleVectors[0], 1)
	assert.Len(t, results, 1)
	assert.Equal(t, uint32(42), results[0].ID)
	assert.InDelta(t, 0.0, results[0].Distance, 1e-6)
}

func TestHNSW_AutoIDEnabled(t *testing.T) {
	t.Parallel()
	dir := createTempDir(t)
	defer deleteDir(t, dir)

	hnsw := hnswgo.New(dir, makeConfig(hnswgo.CosineSpace, true), zerolog.Nop())

	for i, vector := range sampleVectors {
		id, err := hnsw.AddPointAutoID(vector)
		assert.NoError(t, err)
		assert.Equal(t, i+1, int(id))
	}

	assert.Error(t, hnsw.AddPoint(sampleVectors[0], uint32(42)))

	results := hnsw.SearchKNN(sampleVectors[0], 2)
	assert.Equal(t, uint32(1), results[0].ID)
	assert.Equal(t, uint32(2), results[1].ID)
}

func TestHNSW_MarkDelete(t *testing.T) {
	t.Parallel()
	dir := createTempDir(t)
	defer deleteDir(t, dir)

	hnsw := hnswgo.New(dir, makeConfig(hnswgo.CosineSpace, false), zerolog.Nop())

	for i, vector := range sampleVectors {
		require.NoError(t, hnsw.AddPoint(vector, uint32(i)))
	}

	results := hnsw.SearchKNN(sampleVectors[0], 1)
	require.Equal(t, uint32(0), results[0].ID)

	assert.NoError(t, hnsw.MarkDelete(0))

	results = hnsw.SearchKNN(sampleVectors[0], 1)
	assert.Equal(t, uint32(1), results[0].ID)
}

func TestHNSW_SaveAndLoad(t *testing.T) {
	t.Parallel()

	t.Run("load after explicit save", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		var originalResults []hnswgo.KNNResult
		{
			hnsw := hnswgo.New(dir, makeConfig(hnswgo.CosineSpace, false), zerolog.Nop())

			for i, vector := range sampleVectors {
				require.NoError(t, hnsw.AddPoint(vector, uint32(i)))
			}

			originalResults = hnsw.SearchKNN(sampleVectors[0], 2)
			require.Len(t, originalResults, 2)

			assert.NoError(t, hnsw.Save())
		}

		hnsw, err := hnswgo.Load(dir, zerolog.Nop())
		require.NoError(t, err)

		newResults := hnsw.SearchKNN(sampleVectors[0], 2)
		assert.Equal(t, originalResults, newResults)
	})

	t.Run("load auto-id index from log without explicit save", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		{
			hnsw := hnswgo.New(dir, makeConfig(hnswgo.CosineSpace, true), zerolog.Nop())
			// Initial save, just for creating the files
			require.NoError(t, hnsw.Save())

			_, err := hnsw.AddPointAutoID(sampleVectors[0])
			require.NoError(t, err)

			_, err = hnsw.AddPointAutoID(sampleVectors[1])
			require.NoError(t, err)

			require.NoError(t, hnsw.MarkDelete(2))
			require.NoError(t, hnsw.SetEf(150))
		}

		hnsw, err := hnswgo.Load(dir, zerolog.Nop())
		require.NoError(t, err)

		results := hnsw.SearchKNN(sampleVectors[0], 2)
		assert.Len(t, results, 1)
		assert.Equal(t, uint32(1), results[0].ID)
	})

	t.Run("load custom-id index from log without explicit save", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		{
			hnsw := hnswgo.New(dir, makeConfig(hnswgo.CosineSpace, false), zerolog.Nop())
			// Initial save, just for creating the files
			require.NoError(t, hnsw.Save())

			require.NoError(t, hnsw.AddPoint(sampleVectors[0], 1))
			require.NoError(t, hnsw.AddPoint(sampleVectors[1], 2))
			require.NoError(t, hnsw.MarkDelete(2))
			require.NoError(t, hnsw.SetEf(150))
		}

		hnsw, err := hnswgo.Load(dir, zerolog.Nop())
		require.NoError(t, err)

		results := hnsw.SearchKNN(sampleVectors[0], 2)
		assert.Len(t, results, 1)
		assert.Equal(t, uint32(1), results[0].ID)
	})

	t.Run("index is still recovered from a partially corrupted log", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		{
			hnsw := hnswgo.New(dir, makeConfig(hnswgo.CosineSpace, true), zerolog.Nop())
			// Initial save, just for creating the files
			require.NoError(t, hnsw.Save())

			_, err := hnsw.AddPointAutoID(sampleVectors[0])
			require.NoError(t, err)
		}

		file, err := os.OpenFile(path.Join(dir, "log"), os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0666)
		require.NoError(t, err)
		_, err = file.Write([]byte("foo!"))
		require.NoError(t, err)
		require.NoError(t, file.Close())

		hnsw, err := hnswgo.Load(dir, zerolog.Nop())
		require.NoError(t, err)

		results := hnsw.SearchKNN(sampleVectors[0], 1)
		assert.Len(t, results, 1)
		assert.Equal(t, uint32(1), results[0].ID)
	})
}

func TestHNSW_Save(t *testing.T) {
	t.Parallel()

	t.Run("path does not exist", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		hnsw := hnswgo.New(path.Join(dir, "foo", "bar"), makeConfig(hnswgo.CosineSpace, true), zerolog.Nop())
		assert.Error(t, hnsw.Save())
	})
}

func TestHNSW_Loading(t *testing.T) {
	t.Parallel()

	t.Run("state.tmp exists", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		createAndSaveSampleIndex(t, dir)
		createEmptyFile(t, path.Join(dir, "state.tmp"))

		hnsw, err := hnswgo.Load(dir, zerolog.Nop())
		assert.NotNil(t, hnsw)
		assert.NoError(t, err)
	})

	t.Run("state does not exist", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		createAndSaveSampleIndex(t, dir)
		require.NoError(t, os.Remove(path.Join(dir, "state")))

		hnsw, err := hnswgo.Load(dir, zerolog.Nop())
		assert.Nil(t, hnsw)
		assert.Error(t, err)
	})

	t.Run("index.tmp exists", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		createAndSaveSampleIndex(t, dir)
		createEmptyFile(t, path.Join(dir, "index.tmp"))

		hnsw, err := hnswgo.Load(dir, zerolog.Nop())
		assert.NotNil(t, hnsw)
		assert.NoError(t, err)
	})

	t.Run("index does not exist", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		createAndSaveSampleIndex(t, dir)
		require.NoError(t, os.Remove(path.Join(dir, "index")))

		hnsw, err := hnswgo.Load(dir, zerolog.Nop())
		assert.Nil(t, hnsw)
		assert.Error(t, err)
	})

	t.Run("the path does not exist", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)

		hnsw, err := hnswgo.Load(path.Join(dir, "foo"), zerolog.Nop())
		assert.Nil(t, hnsw)
		assert.Error(t, err)
	})
}

func createAndSaveSampleIndex(t *testing.T, dir string) {
	t.Helper()
	hnsw := hnswgo.New(dir, makeConfig(hnswgo.CosineSpace, true), zerolog.Nop())
	for i, vector := range sampleVectors {
		id, err := hnsw.AddPointAutoID(vector)
		require.NoError(t, err)
		require.Equal(t, i+1, int(id))
	}
	require.NoError(t, hnsw.Save())

	require.FileExists(t, path.Join(dir, "state"))
	require.FileExists(t, path.Join(dir, "index"))
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
	t.Helper()
	dir, err := os.MkdirTemp("", "hnsw_test")
	require.NoError(t, err)
	return dir
}

func deleteDir(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.RemoveAll(dir))
}

func createEmptyFile(t *testing.T, name string) {
	t.Helper()
	file, err := os.Create(name)
	require.NoError(t, err)
	require.NoError(t, file.Close())
}
