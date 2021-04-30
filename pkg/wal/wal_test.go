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
		log := wal.NewLog(path.Join(dir, "log"))
		defer mustCloseLog(t, log)

		require.NoError(t, log.WritePointAddition([]float32{1, 2, 3}, 10))
		require.NoError(t, log.WritePointAddition([]float32{4, 5, 6}, 20))
		require.NoError(t, log.WriteDeletionMark(10))
		require.NoError(t, log.WriteEfSetting(42))

		actualEntries := make([]interface{}, 0)
		err := log.Read(func(e interface{}) error {
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
		log := wal.NewLog(path.Join(dir, "foo", "log"))
		defer mustCloseLog(t, log)

		assert.Error(t, log.WriteEfSetting(42))
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
		defer mustCloseLog(t, log)

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
		defer mustCloseLog(t, log)

		file, err := os.Create(filename)
		require.NoError(t, err)
		require.NoError(t, file.Close())

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
		defer mustCloseLog(t, log)

		require.NoError(t, log.WriteEfSetting(1))
		require.NoError(t, log.WriteEfSetting(2))
		require.NoError(t, log.Close())

		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0666)
		require.NoError(t, err)
		_, err = file.Write([]byte("foo!"))
		require.NoError(t, err)
		require.NoError(t, file.Close())

		require.NoError(t, log.WriteEfSetting(3))
		require.NoError(t, log.WriteEfSetting(4))

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
		defer mustCloseLog(t, log)

		file, err := os.Create(filename)
		require.NoError(t, err)
		_, err = file.Write([]byte("foo!"))
		require.NoError(t, err)
		require.NoError(t, file.Close())

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
		log := wal.NewLog(path.Join(dir, "log"))
		defer mustCloseLog(t, log)

		for i := 1; i <= 4; i++ {
			require.NoError(t, log.WriteEfSetting(i))
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

		assert.NoError(t, log.Close())
		assert.NoFileExists(t, filename)
	})

	t.Run("closing after writing", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		log := wal.NewLog(path.Join(dir, "log"))

		require.NoError(t, log.WriteEfSetting(42))
		assert.NoError(t, log.Close())
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
		defer mustCloseLog(t, log)

		require.NoFileExists(t, filename)
		assert.NoError(t, log.Delete())
	})

	t.Run("it removes the file", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		filename := path.Join(dir, "log")
		log := wal.NewLog(filename)
		defer mustCloseLog(t, log)

		file, err := os.Create(filename)
		require.NoError(t, err)
		require.NoError(t, file.Close())

		assert.NoError(t, log.Delete())
		assert.NoFileExists(t, filename)
	})

	t.Run("it removes the file after writing", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		filename := path.Join(dir, "log")
		log := wal.NewLog(filename)
		defer mustCloseLog(t, log)

		require.NoError(t, log.WriteEfSetting(42))

		assert.FileExists(t, filename)
		assert.NoError(t, log.Delete())
		assert.NoFileExists(t, filename)
	})

	t.Run("file existence check error", func(t *testing.T) {
		t.Parallel()
		dir := createTempDir(t)
		defer deleteDir(t, dir)
		log := wal.NewLog(dir) // dir instead of file
		defer mustCloseLog(t, log)

		assert.Error(t, log.Delete())
		assert.DirExists(t, dir, "the dir must not be removed")
	})
}

func mustCloseLog(t *testing.T, l *wal.Log) {
	t.Helper()
	require.NoError(t, l.Close())
}

func createTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "wal_test")
	require.NoError(t, err)
	return dir
}

func deleteDir(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.RemoveAll(dir))
}
