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

package wal_test

import (
	"fmt"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/wal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func TestLog(t *testing.T) {
	t.Parallel()

	t.Run("write and read entries", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		filename := path.Join(dir, "log")
		log := wal.NewLog(filename)
		defer log.Close()

		err := log.WritePointAddition([]float32{1, 2, 3}, 10)
		require.NoError(t, err)

		err = log.WritePointAddition([]float32{4, 5, 6}, 20)
		require.NoError(t, err)

		err = log.WriteDeletionMark(10)
		require.NoError(t, err)

		err = log.WriteEfSetting(42)
		require.NoError(t, err)

		actualEntries := make([]interface{}, 0)
		err = log.Read(func(e interface{}) error {
			actualEntries = append(actualEntries, e)
			return nil
		})
		assert.NoError(t, err)

		expectedEntries := []interface{}{
			wal.PointAddition{Vector: []float32{1, 2, 3}, ID: 10},
			wal.PointAddition{Vector: []float32{4, 5, 6}, ID: 20},
			wal.DeletionMark{ID: 10},
			wal.EfSetting{Ef: 42},
		}
		assert.Equal(t, expectedEntries, actualEntries)
	})

	t.Run("error creating file", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		filename := path.Join(dir, "foo", "log")
		log := wal.NewLog(filename)
		defer log.Close()

		err := log.WriteEfSetting(42)
		assert.Error(t, err)
	})
}

func TestLog_Read(t *testing.T) {
	t.Parallel()

	t.Run("read nonexistent file", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		filename := path.Join(dir, "log")
		log := wal.NewLog(filename)
		defer log.Close()

		require.NoFileExists(t, filename)
		err := log.Read(func(e interface{}) error {
			assert.Fail(t, "read callback should not be invoked")
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("read empty file", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		filename := path.Join(dir, "log")
		log := wal.NewLog(filename)
		defer log.Close()

		file, err := os.Create(filename)
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		err = log.Read(func(e interface{}) error {
			assert.Fail(t, "read callback should not be invoked")
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("read partially corrupted file", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		filename := path.Join(dir, "log")
		log := wal.NewLog(filename)
		defer log.Close()

		err := log.WriteEfSetting(1)
		require.NoError(t, err)
		err = log.WriteEfSetting(2)
		require.NoError(t, err)

		err = log.Close()
		require.NoError(t, err)

		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0666)
		require.NoError(t, err)
		_, err = file.Write([]byte("foo!"))
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		err = log.WriteEfSetting(3)
		require.NoError(t, err)
		err = log.WriteEfSetting(4)
		require.NoError(t, err)

		actualEntries := make([]interface{}, 0)
		err = log.Read(func(e interface{}) error {
			actualEntries = append(actualEntries, e)
			return nil
		})
		assert.Error(t, err)

		expectedEntries := []interface{}{
			wal.EfSetting{Ef: 1},
			wal.EfSetting{Ef: 2},
		}
		assert.Equal(t, expectedEntries, actualEntries)
	})

	t.Run("read corrupted file", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		filename := path.Join(dir, "log")
		log := wal.NewLog(filename)
		defer log.Close()

		file, err := os.Create(filename)
		require.NoError(t, err)
		_, err = file.Write([]byte("foo!"))
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		err = log.Read(func(e interface{}) error {
			assert.Fail(t, "read callback should not be invoked")
			return nil
		})
		assert.Error(t, err)
	})

	t.Run("a callback error stops the reading iteration", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		filename := path.Join(dir, "log")
		log := wal.NewLog(filename)
		defer log.Close()

		for i := 1; i <= 4; i++ {
			err := log.WriteEfSetting(i)
			require.NoError(t, err)
		}

		myError := fmt.Errorf("my error")

		actualEntries := make([]interface{}, 0)
		err := log.Read(func(e interface{}) error {
			if len(actualEntries) == 2 {
				return myError
			}
			actualEntries = append(actualEntries, e)
			return nil
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, myError)

		expectedEntries := []interface{}{
			wal.EfSetting{Ef: 1},
			wal.EfSetting{Ef: 2},
		}
		assert.Equal(t, expectedEntries, actualEntries)
	})
}

func TestLog_Close(t *testing.T) {
	t.Parallel()

	t.Run("closing without prior writing", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		filename := path.Join(dir, "log")
		log := wal.NewLog(filename)

		err := log.Close()
		assert.NoError(t, err)
		assert.NoFileExists(t, filename)
	})

	t.Run("closing after writing", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		filename := path.Join(dir, "log")
		log := wal.NewLog(filename)

		err := log.WriteEfSetting(42)
		require.NoError(t, err)

		err = log.Close()
		assert.NoError(t, err)
	})
}

func TestLog_Delete(t *testing.T) {
	t.Parallel()

	t.Run("it does nothing if file does not exist", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		filename := path.Join(dir, "log")
		log := wal.NewLog(filename)
		defer log.Close()

		require.NoFileExists(t, filename)
		err := log.Delete()
		assert.NoError(t, err)
	})

	t.Run("it removes the file", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		filename := path.Join(dir, "log")
		log := wal.NewLog(filename)
		defer log.Close()

		file, err := os.Create(filename)
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		err = log.Delete()
		assert.NoError(t, err)
		assert.NoFileExists(t, filename)
	})

	t.Run("it removes the file after writing", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		filename := path.Join(dir, "log")
		log := wal.NewLog(filename)
		defer log.Close()

		err := log.WriteEfSetting(42)
		require.NoError(t, err)

		assert.FileExists(t, filename)
		err = log.Delete()
		assert.NoError(t, err)
		assert.NoFileExists(t, filename)
	})

	t.Run("file existence check error", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		log := wal.NewLog(dir) // dir instead of file
		defer log.Close()

		err := log.Delete()
		assert.Error(t, err)
		assert.DirExists(t, dir, "the dir must not be removed")
	})
}

func createTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "wal_test")
	require.NoError(t, err)
	return dir
}

func deleteDir(t *testing.T, dir string) {
	t.Helper()
	err := os.RemoveAll(dir)
	require.NoError(t, err)
}
