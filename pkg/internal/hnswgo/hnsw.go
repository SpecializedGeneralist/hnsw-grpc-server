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
	"fmt"
	"math"
	"sync/atomic"
	"unsafe"
)

type HNSW struct {
	index     C.HNSW
	SpaceType string
	Dim       int
	LastID    uint32
	AutoID    bool
}

func New(dim, M, efConstruction, randSeed int, maxElements uint32, spaceType string, autoID bool) *HNSW {
	var hnsw HNSW
	hnsw.LastID = 0
	hnsw.Dim = dim
	hnsw.SpaceType = spaceType
	hnsw.AutoID = autoID
	switch spaceType {
	case "ip", "cosine":
		hnsw.index = C.initHNSW(
			C.int(dim),
			C.ulong(maxElements),
			C.int(M),
			C.int(efConstruction),
			C.int(randSeed),
			C.char('i'),
		)
	default:
		hnsw.index = C.initHNSW(
			C.int(dim),
			C.ulong(maxElements),
			C.int(M),
			C.int(efConstruction),
			C.int(randSeed),
			C.char('l'),
		)
	}
	return &hnsw
}

func Load(location string, dim int, spaceType string, autoID string) *HNSW {
	var hnsw HNSW
	hnsw.Dim = dim
	hnsw.SpaceType = spaceType
	hnsw.LastID = 0
	pLocation := C.CString(location)
	hnsw.AutoID = autoID == "true"
	switch spaceType {
	case "ip", "cosine":
		hnsw.index = C.loadHNSW(pLocation, C.int(dim), C.char('i'))
	default:
		hnsw.index = C.loadHNSW(pLocation, C.int(dim), C.char('l'))
	}
	C.free(unsafe.Pointer(pLocation))
	return &hnsw
}

func (h *HNSW) Save(location string) {
	pLocation := C.CString(location)
	C.saveHNSW(h.index, pLocation)
	C.free(unsafe.Pointer(pLocation))
}

func normalizeVector(vector []float32) []float32 {
	var norm float32
	for i := 0; i < len(vector); i++ {
		norm += vector[i] * vector[i]
	}
	norm = 1.0 / (float32(math.Sqrt(float64(norm))) + 1e-15)
	for i := 0; i < len(vector); i++ {
		vector[i] = vector[i] * norm
	}
	return vector
}

func (h *HNSW) AddPointAutoID(vector []float32) (uint32, error) {
	if !h.AutoID {
		return 0, fmt.Errorf("invalid call with auto-id disabled")
	}
	id := atomic.AddUint32(&h.LastID, 1)
	if h.SpaceType == "cosine" {
		vector = normalizeVector(vector)
	}
	C.addPoint(h.index, (*C.float)(unsafe.Pointer(&vector[0])), C.ulong(id))
	return id, nil
}

func (h *HNSW) AddPoint(vector []float32, label uint32) error {
	if h.AutoID {
		return fmt.Errorf("invalid call with auto-id enabled")
	}
	if h.SpaceType == "cosine" {
		vector = normalizeVector(vector)
	}
	C.addPoint(h.index, (*C.float)(unsafe.Pointer(&vector[0])), C.ulong(label))
	return nil
}

func (h *HNSW) MarkDelete(label uint32) {
	C.markDelete(h.index, C.ulong(label))
}

func (h *HNSW) SearchKNN(vector []float32, N int) ([]uint32, []float32) {
	Clabel := make([]C.ulong, N, N)
	Cdist := make([]C.float, N, N)
	if h.SpaceType == "cosine" {
		vector = normalizeVector(vector)
	}
	numResult := int(C.searchKnn(
		h.index, (*C.float)(unsafe.Pointer(&vector[0])),
		C.int(N),
		&Clabel[0],
		&Cdist[0]),
	)
	labels := make([]uint32, N)
	dists := make([]float32, N)
	for i := 0; i < numResult; i++ {
		labels[i] = uint32(Clabel[i])
		dists[i] = float32(Cdist[i])
	}
	return labels[:numResult], dists[:numResult]
}

func (h *HNSW) SetEf(ef int) {
	C.setEf(h.index, C.int(ef))
}
