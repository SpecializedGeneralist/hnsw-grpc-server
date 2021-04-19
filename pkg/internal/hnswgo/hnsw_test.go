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
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/internal/hnswgo"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var sampleVectors = [][]float32{
	{0.1, 0.2, 0.3, 0.4, 0.5},
	{0.9, 0.8, 0.7, 0.6, 0.5},
}

func TestSpaceTypeFromString(t *testing.T) {
	t.Run("IPSpace", func(t *testing.T) {
		val, err := hnswgo.SpaceTypeFromString("ip")
		assert.NoError(t, err)
		assert.Equal(t, hnswgo.IPSpace, val)
	})
	t.Run("CosineSpace", func(t *testing.T) {
		val, err := hnswgo.SpaceTypeFromString("cosine")
		assert.NoError(t, err)
		assert.Equal(t, hnswgo.CosineSpace, val)
	})
	t.Run("L2Space", func(t *testing.T) {
		val, err := hnswgo.SpaceTypeFromString("l2")
		assert.NoError(t, err)
		assert.Equal(t, hnswgo.L2Space, val)
	})
	t.Run("invalid", func(t *testing.T) {
		_, err := hnswgo.SpaceTypeFromString("foo")
		assert.Error(t, err)
	})
}

func TestHNSW_IPSpace(t *testing.T) {
	hnsw := hnswgo.New(hnswgo.Config{
		SpaceType:      hnswgo.IPSpace,
		Dim:            5,
		MaxElements:    10,
		M:              10,
		EfConstruction: 200,
		RandSeed:       100,
		AutoIDEnabled:  false,
	})

	for i, vector := range sampleVectors {
		err := hnsw.AddPoint(vector, uint32(i))
		assert.NoError(t, err)
	}

	results := hnsw.SearchKNN(sampleVectors[0], 2)
	assert.Len(t, results, 2)
}

func TestHNSW_L2Space(t *testing.T) {
	hnsw := hnswgo.New(hnswgo.Config{
		SpaceType:      hnswgo.L2Space,
		Dim:            5,
		MaxElements:    10,
		M:              10,
		EfConstruction: 200,
		RandSeed:       100,
		AutoIDEnabled:  false,
	})

	for i, vector := range sampleVectors {
		err := hnsw.AddPoint(vector, uint32(i))
		assert.NoError(t, err)
	}

	results := hnsw.SearchKNN(sampleVectors[0], 2)
	assert.Len(t, results, 2)
}

func TestHNSW_CosineSpace(t *testing.T) {
	hnsw := hnswgo.New(hnswgo.Config{
		SpaceType:      hnswgo.CosineSpace,
		Dim:            5,
		MaxElements:    10,
		M:              10,
		EfConstruction: 200,
		RandSeed:       100,
		AutoIDEnabled:  false,
	})

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
	hnsw := hnswgo.New(hnswgo.Config{
		SpaceType:      hnswgo.CosineSpace,
		Dim:            5,
		MaxElements:    10,
		M:              10,
		EfConstruction: 200,
		RandSeed:       100,
		AutoIDEnabled:  false,
	})

	err := hnsw.AddPoint(sampleVectors[0], uint32(42))
	assert.NoError(t, err)

	err = hnsw.AddPointAutoID(sampleVectors[1])
	assert.Error(t, err)

	results := hnsw.SearchKNN(sampleVectors[0], 1)
	assert.Len(t, results, 1)
	assert.Equal(t, uint32(42), results[0].ID)
	assert.InDelta(t, 0.0, results[0].Distance, 1e-6)
}

func TestHNSW_AutoIDEnabled(t *testing.T) {
	hnsw := hnswgo.New(hnswgo.Config{
		SpaceType:      hnswgo.CosineSpace,
		Dim:            5,
		MaxElements:    10,
		M:              10,
		EfConstruction: 200,
		RandSeed:       100,
		AutoIDEnabled:  true,
	})

	for _, vector := range sampleVectors {
		err := hnsw.AddPointAutoID(vector)
		assert.NoError(t, err)
	}

	err := hnsw.AddPoint(sampleVectors[0], uint32(42))
	assert.Error(t, err)

	results := hnsw.SearchKNN(sampleVectors[0], 2)
	assert.Equal(t, uint32(1), results[0].ID)
	assert.Equal(t, uint32(2), results[1].ID)
}

func TestHNSW_MarkDelete(t *testing.T) {
	hnsw := hnswgo.New(hnswgo.Config{
		SpaceType:      hnswgo.CosineSpace,
		Dim:            5,
		MaxElements:    10,
		M:              10,
		EfConstruction: 200,
		RandSeed:       100,
		AutoIDEnabled:  false,
	})

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
	tmpPath, err := os.MkdirTemp("", "hnsw_test")
	assert.NoError(t, err)
	defer func() {
		err = os.RemoveAll(tmpPath)
		assert.NoError(t, err)
	}()

	var originalResults []hnswgo.KNNResult

	{
		hnsw := hnswgo.New(hnswgo.Config{
			SpaceType:      hnswgo.CosineSpace,
			Dim:            5,
			MaxElements:    10,
			M:              10,
			EfConstruction: 200,
			RandSeed:       100,
			AutoIDEnabled:  false,
		})

		for i, vector := range sampleVectors {
			err = hnsw.AddPoint(vector, uint32(i))
			assert.NoError(t, err)
		}

		originalResults = hnsw.SearchKNN(sampleVectors[0], 2)
		assert.Len(t, originalResults, 2)

		err = hnsw.Save(tmpPath)
		assert.NoError(t, err)
	}

	hnsw, err := hnswgo.Load(tmpPath)
	assert.NoError(t, err)

	newResults := hnsw.SearchKNN(sampleVectors[0], 2)
	assert.Equal(t, originalResults, newResults)
}
