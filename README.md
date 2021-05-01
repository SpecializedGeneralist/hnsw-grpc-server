# hnsw-grpc-server

[![Go Report Card](https://goreportcard.com/badge/github.com/SpecializedGeneralist/hnsw-grpc-server)](https://goreportcard.com/report/github.com/SpecializedGeneralist/hnsw-grpc-server)
[![Maintainability](https://api.codeclimate.com/v1/badges/7b5c7fd17aada5ad4016/maintainability)](https://codeclimate.com/github/SpecializedGeneralist/hnsw-grpc-server/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/7b5c7fd17aada5ad4016/test_coverage)](https://codeclimate.com/github/SpecializedGeneralist/hnsw-grpc-server/test_coverage)

This is a gRPC server for [hnswlib](https://github.com/nmslib/hnswlib).

It provides more than just the core HNSW model: it is a tool that can be used
end-to-end, supporting TLS encryption, multiple persistent indices and batch
insertions.

This repository includes the relevant sources from the [hnswlib](https://github.com/nmslib/hnswlib),
so it doesn't require any external dependency. For more information please
refer to [hnswlib](https://github.com/nmslib/hnswlib) and
[Efficient and robust approximate nearest neighbor search using Hierarchical
Navigable Small World graphs](https://arxiv.org/abs/1603.09320).

## Features

A list of API methods now follows:

| Method | Description |
| -------------- | --------- |
| CreateIndex | Make a new index |
| DeleteIndex | Removes an index |
| InsertVector | Insert a new vector in the given index, letting the index generate an ID |
| InsertVectors | Insert new vectors in the given index, with generated ID, then flush the index |
| InsertVectorWithId | Insert a new vector with given ID in the given index |
| InsertVectorsWithId | Insert new vectors with given IDs in the given index, then flush the index |
| SearchKNN | Return the top k nearest neighbors to the query, searching on the given index |
| FlushIndex | Serialize the index to file |
| Indices | Return the list of indices |
| SetEf | Set the `ef` parameter for the given index |

## Build and run

You first need to compile the C++ `hnswlib` wrapper. Just run the following
script to compile it with `g++`:

```shell
./pkg/hnswgo/make.sh
```

Then, download Go dependencies and build the package (with `cgo` enabled):

```shell
go mod download
CGO_CXXFLAGS="-std=c++11" CGO_ENABLED=1 go build \
  -ldflags="-extldflags=-static" \
  -o hnsw-grpc-server \
  cmd/main.go
```

You can finally run the executable. For example, you can get help running:

```shell
./hnsw-grpc-server -h
```

## Docker

The [Docker](https://www.docker.com/) image can be built like this:

```console
docker build -t hnsw-grpc-server:latest .
```

Pre-built images are available on Docker Hub, at [specializedgeneralist/hnsw-grpc-server](https://hub.docker.com/r/specializedgeneralist/hnsw-grpc-server).

For example, you can pull the image and run the server like this:

```shell
docker run -d \
    --name hnsw-grpc-server \
    -v /path/to/your/data/folder:/hnsw-grpc-server-data
    -p 19530:19530 \
    specializedgeneralist/hnsw-grpc-server:1.0.0
```

## Credits

- [hnswlib](https://github.com/nmslib/hnswlib) - Header-only C++ HNSW implementation.
- [hnswgo](https://github.com/evan176/hnswgo) - Go interface for hnswlib.
