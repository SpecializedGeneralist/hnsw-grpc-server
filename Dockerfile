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

FROM golang:1.15.7-alpine3.13 as Builder

RUN set -eux; \
	apk add --no-cache --virtual .build-deps \
		gcc \
		g++ \
		musl-dev \
        ;

RUN mkdir /data
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN cd pkg/internal/hnswgo/ && sh make.sh
RUN go mod download
RUN GOOS=linux GOARCH=amd64 CGO_CXXFLAGS="-std=c++11" CGO_ENABLED=1 go build -ldflags="-extldflags=-static" -o entrypoint cmd/main.go

FROM scratch

COPY --from=Builder /data /data
COPY --from=Builder /build/entrypoint /entrypoint

ENV GOOS linux
ENV GOARCH amd64
ENTRYPOINT ["/entrypoint"]
CMD ["help"]
