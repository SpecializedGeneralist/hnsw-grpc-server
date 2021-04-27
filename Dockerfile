# Copyright 2021 SpecializedGeneralist
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.16.3-alpine3.13 as Builder

RUN apk add --no-cache \
        gcc \
        g++ \
        musl-dev

WORKDIR /go/src/hnsw-grpc-server
COPY . .

RUN set -x \
    && mkdir /usr/local/hnsw-grpc-server-data \
    && pkg/hnswgo/make.sh \
    && go mod download \
    && CGO_CXXFLAGS="-std=c++11" CGO_ENABLED=1 go build \
        -ldflags="-extldflags=-static" \
        -o /go/bin/hnsw-grpc-server \
        cmd/main.go

FROM scratch

COPY --from=Builder /go/bin/hnsw-grpc-server /hnsw-grpc-server
COPY --from=Builder /usr/local/hnsw-grpc-server-data /hnsw-grpc-server-data

ENTRYPOINT ["/hnsw-grpc-server"]
