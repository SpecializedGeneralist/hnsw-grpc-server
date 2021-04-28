package hnswgo

// #cgo LDFLAGS: -L${SRCDIR}/hnsw -lm
// #include <stdlib.h>
// #include "hnsw_wrapper.h"
// HNSW initHNSW(int dim, unsigned long int max_elements, int M, int ef_construction, int rand_seed, char stype);
// HNSW loadHNSW(char *location, int dim, char stype);
// void saveHNSW(HNSW index, char *location);
// void addPoint(HNSW index, float *vec, unsigned long int label);
// void markDelete(HNSW index, unsigned long int label);
// int searchKnn(HNSW index, float *vec, int N, unsigned long int *label, float *dist);
// void setEf(HNSW index, int ef);
import "C"

import (
	"encoding/gob"
	"fmt"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/osutils"
	"math"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Config provides configuration parameters for HNSW.
type Config struct {
	SpaceType      SpaceType
	Dim            int
	MaxElements    int
	M              int
	EfConstruction int
	RandSeed       int
	AutoIDEnabled  bool
}

// SpaceType identifies a space type to be used by HNSW algorithm.
type SpaceType string

// HNSW is an interface to HNSW C code.
type HNSW struct {
	dir   string
	index C.HNSW
	state hnswState
	// Most operations lock the mutex for reading, including AddPoint and
	// AddPointAutoID, since the actual locking of critical parts is
	// already implemented in the native C++ code.
	// The only operation which locks for writing is Save.
	rwMx sync.RWMutex
}

// hnswState provides serializable configuration settings and other
// parameters for the internal state of a HNSW object.
type hnswState struct {
	Config
	LastAutoID uint32
}

const (
	// IPSpace identifies an Inner Product space.
	IPSpace SpaceType = "ip"
	// CosineSpace identifies a Cosine space.
	CosineSpace SpaceType = "cosine"
	// L2Space identifies an L2 space.
	L2Space SpaceType = "l2"
)

// SpaceTypeFromString makes a SpaceType value from string.
// Valid string values are: "ip", "cosine", or "l2".
func SpaceTypeFromString(s string) (SpaceType, error) {
	switch s {
	case "ip":
		return IPSpace, nil
	case "cosine":
		return CosineSpace, nil
	case "l2":
		return L2Space, nil
	default:
		return IPSpace, fmt.Errorf("invalid space type %#v", s)
	}
}

func (st SpaceType) cChar() C.char {
	switch st {
	case IPSpace, CosineSpace:
		return C.char('i')
	case L2Space:
		return C.char('l')
	default:
		panic(fmt.Sprintf("unexpected SpaceType %#v", st))
	}
}

// New creates a new HNSW index.
func New(dir string, config Config) *HNSW {
	return &HNSW{
		dir: dir,
		index: C.initHNSW(
			C.int(config.Dim),
			C.ulong(config.MaxElements),
			C.int(config.M),
			C.int(config.EfConstruction),
			C.int(config.RandSeed),
			config.SpaceType.cChar(),
		),
		state: hnswState{
			Config:     config,
			LastAutoID: 0,
		},
		rwMx: sync.RWMutex{},
	}
}

// Load loads an HNSW index from file.
func Load(dir string) (*HNSW, error) {
	state, err := loadState(dir)
	if err != nil {
		return nil, err
	}

	index, err := loadIndex(dir, state.Dim, state.SpaceType)
	if err != nil {
		return nil, err
	}

	return &HNSW{
		dir:   dir,
		index: index,
		state: *state,
		rwMx:  sync.RWMutex{},
	}, nil
}

func loadState(dir string) (_ *hnswState, err error) {
	tmpFilename := path.Join(dir, "state.tmp")
	tmpExists, err := osutils.FileExists(tmpFilename)
	if err != nil {
		return nil, err
	}
	if tmpExists {
		return nil, fmt.Errorf("cannot load HNSW state file: %#v found", tmpFilename)
	}

	filename := path.Join(dir, "state")
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file %#v: %w", filename, err)
	}
	defer func() {
		if e := file.Close(); e != nil && err == nil {
			err = fmt.Errorf("error closing file %#v: %w", filename, e)
		}
	}()
	decoder := gob.NewDecoder(file)
	state := new(hnswState)
	err = decoder.Decode(state)
	if err != nil {
		return nil, fmt.Errorf("error encoding HNSW state: %w", err)
	}
	return state, nil
}

func loadIndex(dir string, dim int, spaceType SpaceType) (C.HNSW, error) {
	tmpFilename := path.Join(dir, "index.tmp")
	tmpExists, err := osutils.FileExists(tmpFilename)
	if err != nil {
		return nil, err
	}
	if tmpExists {
		return nil, fmt.Errorf("cannot load HNSW index file: %#v found", tmpFilename)
	}

	filename := path.Join(dir, "index")
	fExists, err := osutils.FileExists(filename)
	if err != nil {
		return nil, err
	}
	if !fExists {
		return nil, fmt.Errorf("cannot load HNSW index file %#v: file not found", filename)
	}

	pFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(pFilename))
	index := C.loadHNSW(
		pFilename,
		C.int(dim),
		spaceType.cChar(),
	)
	return index, nil
}

// Save saves the HNSW index to file.
func (h *HNSW) Save() error {
	h.rwMx.Lock()
	defer h.rwMx.Unlock()

	err := ensureDirExists(h.dir)
	if err != nil {
		return err
	}

	// Create new temporary files: if something goes wrong, the old
	// files (if any) will not be corrupted.
	err = h.saveState(path.Join(h.dir, "state.tmp"))
	if err != nil {
		return err
	}
	h.saveIndex(path.Join(h.dir, "index.tmp"))

	// Now that the temporary files are successfully created, replace
	// the old files (if any) with the new ones.
	err = os.Rename(path.Join(h.dir, "state.tmp"), path.Join(h.dir, "state"))
	if err != nil {
		return err
	}
	err = os.Rename(path.Join(h.dir, "index.tmp"), path.Join(h.dir, "index"))
	if err != nil {
		return err
	}

	return nil
}

func (h *HNSW) saveState(name string) (err error) {
	file, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("error creating file %#v: %w", name, err)
	}
	defer func() {
		if e := file.Close(); e != nil && err == nil {
			err = fmt.Errorf("error closing file %#v: %w", name, e)
		}
	}()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(h.state)
	if err != nil {
		return fmt.Errorf("error encoding HNSW state: %w", err)
	}
	return nil
}

func (h *HNSW) saveIndex(name string) {
	pName := C.CString(name)
	defer C.free(unsafe.Pointer(pName))
	C.saveHNSW(h.index, pName)
}

// AddPoint adds a new vector to the index.
func (h *HNSW) AddPoint(vector []float32, id uint32) error {
	if h.state.AutoIDEnabled {
		return fmt.Errorf("invalid call to HNSW.AddPoint with auto-ID enabled")
	}
	h.addPoint(vector, id)
	return nil
}

// AddPointAutoID adds a new vector to the index.
func (h *HNSW) AddPointAutoID(vector []float32) (uint32, error) {
	if !h.state.AutoIDEnabled {
		return 0, fmt.Errorf("invalid call to HNSW.AddPointAutoID with auto-ID disabled")
	}
	id := atomic.AddUint32(&h.state.LastAutoID, 1)
	h.addPoint(vector, id)
	return id, nil
}

func (h *HNSW) addPoint(vector []float32, id uint32) {
	h.rwMx.RLock()
	defer h.rwMx.RUnlock()

	if h.state.SpaceType == "cosine" {
		vector = normalizeVector(vector)
	}
	C.addPoint(h.index, (*C.float)(unsafe.Pointer(&vector[0])), C.ulong(id))
}

// MarkDelete marks an element with the given ID deleted.
// It does not really change the current graph.
func (h *HNSW) MarkDelete(id uint32) {
	h.rwMx.RLock()
	defer h.rwMx.RUnlock()

	C.markDelete(h.index, C.ulong(id))
}

// KNNResult is an ID/Distance pair, which is a single result
// item of HNSW.SearchKNN.
type KNNResult struct {
	ID       uint32
	Distance float32
}

// SearchKNN performs KNN search.
func (h *HNSW) SearchKNN(vector []float32, N int) []KNNResult {
	h.rwMx.RLock()
	defer h.rwMx.RUnlock()

	if h.state.SpaceType == "cosine" {
		vector = normalizeVector(vector)
	}

	cLabels := make([]C.ulong, N, N)
	cDistances := make([]C.float, N, N)
	numResults := int(C.searchKnn(
		h.index,
		(*C.float)(unsafe.Pointer(&vector[0])),
		C.int(N),
		&cLabels[0],
		&cDistances[0],
	))

	results := make([]KNNResult, numResults)
	for i := range results {
		results[i] = KNNResult{
			ID:       uint32(cLabels[i]),
			Distance: float32(cDistances[i]),
		}
	}
	return results
}

// SetEf sets the "ef" parameter.
func (h *HNSW) SetEf(ef int) {
	h.rwMx.RLock()
	defer h.rwMx.RUnlock()

	C.setEf(h.index, C.int(ef))
}

func normalizeVector(vector []float32) []float32 {
	var norm float32
	for _, v := range vector {
		norm += v * v
	}
	norm = 1.0 / (float32(math.Sqrt(float64(norm))) + 1e-15)

	result := make([]float32, len(vector))
	for i := range result {
		result[i] = vector[i] * norm
	}
	return result
}

func ensureDirExists(name string) error {
	exists, err := osutils.DirExists(name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	err = os.Mkdir(name, 0777)
	if err != nil {
		return fmt.Errorf("error creating dir %#v: %w", name, err)
	}
	return nil
}
