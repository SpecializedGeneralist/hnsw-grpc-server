#!/usr/bin/env sh

set -e

cd "$(dirname "$0")"
g++ -O3 -funroll-loops -pthread -std=c++0x -march=native -std=c++11 -I. -c hnsw_wrapper.cc
