# SPDX-License-Identifier: Apache-2.0
# Copyright 2023 Authors of Nimbus

FROM golang:1.22 AS builder
ARG TARGETOS
ARG TARGETARCH

# Required to embed build info into binary.
COPY .git /.git

WORKDIR /nimbus

# relative deps requried by the adapter
ADD api/ api/
ADD pkg/ pkg/
ADD go.mod go.mod
ADD go.sum go.sum

ARG ADAPTER_DIR=pkg/adapter/nimbus-k8tls
WORKDIR /nimbus/$ADAPTER_DIR

COPY $ADAPTER_DIR/go.mod go.mod
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY $ADAPTER_DIR/manager manager
COPY $ADAPTER_DIR/builder builder
COPY $ADAPTER_DIR/watcher watcher
COPY $ADAPTER_DIR/main.go main.go
COPY $ADAPTER_DIR/Makefile Makefile

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} make build

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /nimbus/pkg/adapter/nimbus-k8tls/bin/nimbus-k8tls .
USER 65532:65532

ENTRYPOINT ["/nimbus-k8tls"]
