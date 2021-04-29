#include <iostream>
#include "hnswlib/hnswlib.h"
#include "hnsw_wrapper.h"
#include <thread>
#include <atomic>

HNSW initHNSW(int dim, unsigned long int max_elements, int M, int ef_construction, int rand_seed, char stype) {
  hnswlib::SpaceInterface<float> *space;
  if (stype == 'i') {
    space = new hnswlib::InnerProductSpace(dim);
  } else {
    space = new hnswlib::L2Space(dim);
  }
  hnswlib::HierarchicalNSW<float> *appr_alg = new hnswlib::HierarchicalNSW<float>(space, max_elements, M, ef_construction, rand_seed);
  return (void*)appr_alg;
}

HNSW loadHNSW(char *location, int dim, unsigned long int max_elements, char stype) {
  hnswlib::SpaceInterface<float> *space;
  if (stype == 'i') {
    space = new hnswlib::InnerProductSpace(dim);
  } else {
    space = new hnswlib::L2Space(dim);
  }
  hnswlib::HierarchicalNSW<float> *appr_alg = new hnswlib::HierarchicalNSW<float>(space, std::string(location), false, max_elements);
  return (void*)appr_alg;
}

void saveHNSW(HNSW index, char *location) {
  ((hnswlib::HierarchicalNSW<float>*)index)->saveIndex(location);
}

void addPoint(HNSW index, float *vec, unsigned long int label) {
  ((hnswlib::HierarchicalNSW<float>*)index)->addPoint(vec, label);
}

void markDelete(HNSW index, unsigned long int label) {
  ((hnswlib::HierarchicalNSW<float>*)index)->markDelete(label);
}

int searchKnn(HNSW index, float *vec, int N, unsigned long int *label, float *dist) {
  std::priority_queue<std::pair<float, hnswlib::labeltype>> gt;
  try {
    gt = ((hnswlib::HierarchicalNSW<float>*)index)->searchKnn(vec, N);
  } catch (const std::exception& e) { 
    return 0;
  }

  int n = gt.size();
  std::pair<float, hnswlib::labeltype> pair;
  for (int i = n - 1; i >= 0; i--) {
    pair = gt.top();
    *(dist+i) = pair.first;
    *(label+i) = pair.second;
    gt.pop();
  }
  return n;
}

void setEf(HNSW index, int ef) {
    ((hnswlib::HierarchicalNSW<float>*)index)->ef_ = ef;
}
