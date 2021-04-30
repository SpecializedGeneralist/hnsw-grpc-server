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

// Package wal provides a simple interface for Write-Ahead Logging (WAL) of
// operations on HNSW indices.
package wal

import (
	"encoding/gob"
	"fmt"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/osutils"
	"io"
	"os"
	"sync"
)

// Log is an object to conveniently handle write-ahead log files for
// HNSW indices.
type Log struct {
	filename string
	file     *os.File
	encoder  *gob.Encoder
	mx       sync.Mutex
}

// PointAddition is a log entry representing the operation of adding new data.
type PointAddition struct {
	Vector []float32
	ID     uint32
}

// DeletionMark is a log entry representing the operation of marking data
// for deletion.
type DeletionMark struct {
	ID uint32
}

// EfSetting is a log entry representing the operation of setting the
// "ef" parameter.
type EfSetting struct {
	Ef int
}

func init() {
	gob.Register(PointAddition{})
	gob.Register(DeletionMark{})
	gob.Register(EfSetting{})
}

// NewLog creates a new Log.
func NewLog(filename string) *Log {
	return &Log{
		filename: filename,
		mx:       sync.Mutex{},
	}
}

// WritePointAddition appends a new PointAddition entry to the log.
func (log *Log) WritePointAddition(vector []float32, id uint32) error {
	return log.write(PointAddition{
		Vector: vector,
		ID:     id,
	})
}

// WriteDeletionMark appends a new DeletionMark entry to the log.
func (log *Log) WriteDeletionMark(id uint32) error {
	return log.write(DeletionMark{
		ID: id,
	})
}

// WriteEfSetting appends a new EfSetting entry to the log.
func (log *Log) WriteEfSetting(ef int) error {
	return log.write(EfSetting{
		Ef: ef,
	})
}

// Read reads all entries from the log file and calls the given function for
// each of them.
//
// If the log file does not exists, no error is returned, since this scenario
// is considered an equivalent of having an empty log.
//
// If the log file was still open for writing, it is first closed.
//
// The callback function can return an error; if it is not nil, the entries
// iteration will stop, and the Read function will returned the same error.
func (log *Log) Read(fn func(e interface{}) error) (err error) {
	log.mx.Lock()
	defer log.mx.Unlock()

	if log.file != nil {
		err = log.removeEncoderAndCloseFile()
		if err != nil {
			return err
		}
	}

	file, err := os.Open(log.filename)
	if os.IsNotExist(err) {
		// Equivalent of an empty log.
		return nil
	}
	if err != nil {
		return fmt.Errorf("error opening log file %#v for reading: %w", log.filename, err)
	}
	defer func() {
		if e := file.Close(); e != nil && err == nil {
			err = e
		}
	}()

	return log.readFile(file, fn)
}

func (log *Log) readFile(file *os.File, fn func(e interface{}) error) (err error) {
	decoder := gob.NewDecoder(file)
	for {
		var e interface{}
		err = decoder.Decode(&e)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error decoding entry from log file %#v: %w", log.filename, err)
		}
		err = fn(e)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete removes the log file, virtually emptying the log.
func (log *Log) Delete() error {
	log.mx.Lock()
	defer log.mx.Unlock()

	if log.file != nil {
		err := log.removeEncoderAndCloseFile()
		if err != nil {
			return err
		}
	}

	logExists, err := osutils.FileExists(log.filename)
	if err != nil {
		return err
	}
	if !logExists {
		return nil
	}
	err = os.Remove(log.filename)
	if err != nil {
		return fmt.Errorf("error removing log file %#v: %w", log.filename, err)
	}
	return nil
}

// Close closes the file if necessary, and frees related internal resources.
func (log *Log) Close() error {
	log.mx.Lock()
	defer log.mx.Unlock()

	if log.file == nil {
		return nil
	}
	return log.removeEncoderAndCloseFile()
}

func (log *Log) write(e interface{}) error {
	log.mx.Lock()
	defer log.mx.Unlock()

	if log.file == nil {
		err := log.openFileAndCreateEncoder()
		if err != nil {
			return err
		}
	}

	err := log.encoder.Encode(&e)
	if err != nil {
		return fmt.Errorf("error encoding value %#v to log file %#v: %w", e, log.filename, err)
	}
	err = log.file.Sync()
	if err != nil {
		return fmt.Errorf("error syncing log file %#v: %w", log.filename, err)
	}
	return nil
}

func (log *Log) openFileAndCreateEncoder() error {
	file, err := os.OpenFile(log.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0666)
	if err != nil {
		return fmt.Errorf("error opening log file %#v: %w", log.filename, err)
	}
	log.file = file
	log.encoder = gob.NewEncoder(log.file)
	return nil
}

func (log *Log) removeEncoderAndCloseFile() error {
	err := log.file.Close()
	log.file = nil
	log.encoder = nil
	if err != nil {
		return fmt.Errorf("error closing log file %#v: %w", log.filename, err)
	}
	return nil
}
