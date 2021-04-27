# hnsw-grpc-server

[![Go Report Card](https://goreportcard.com/badge/github.com/SpecializedGeneralist/hnsw-grpc-server)](https://goreportcard.com/report/github.com/SpecializedGeneralist/hnsw-grpc-server)
[![Maintainability](https://api.codeclimate.com/v1/badges/7b5c7fd17aada5ad4016/maintainability)](https://codeclimate.com/github/SpecializedGeneralist/hnsw-grpc-server/maintainability)

This is a work-in-progress gRPC server for [hnswlib](https://github.com/nmslib/hnswlib). 

It is more than just the core HNSW model, it provides a tool that can be used end-to-end, supporting TLS encryption, multiple persistent indices and batch insertions.

This package includes the relevant sources from the [hnswlib](https://github.com/nmslib/hnswlib), so it doesn't require any external dependencies. For more information, please follow [hnswlib](https://github.com/nmslib/hnswlib) and [Efficient and robust approximate nearest neighbor search using Hierarchical Navigable Small World graphs](https://arxiv.org/abs/1603.09320).

## Features

A list of API methods now follows:

| Method | Description | 
| -------------- | --------- |
| CreateIndex | Make a new index |
| DeleteIndex | Removes an index |
| InsertVector | Insert a new vector in the given index |
| InsertVectors | Insert the new vectors in the given index and flush the index |
| SearchKNN | Return the top k nearest neighbors to the query, searching on the given index |
| FlushIndex | Serialize the index to file |
| Indices | Return the list of indices |
| SetEf | Set the `ef` parameter for the given index |


## Usage

### Build

The [Docker](https://www.docker.com/) image can be built like this:

```console
docker build -t hnsw-grpc-server:latest . -f Dockerfile
```

### Run

The container can be run like this:

```console
docker run --network=host --name hnsw-grpc-server -it hnsw-grpc-server:latest --address=0.0.0.0:19530 --debug
```

It should print:

```console
2021-04-04T21:55:37Z INF Starting: gRPC Listener [0.0.0.0:19530]
```

# Credits

- [hnswlib](https://github.com/nmslib/hnswlib) - Header-only C++ HNSW implementation.
- [hnswgo](https://github.com/evan176/hnswgo) - Go interface for hnswlib.
