# SPDX-License-Identifier: Apache-2.0
# Copyright 2023 Authors of Nimbus

FROM golang:1.22 AS builder
ARG TARGETOS
ARG TARGETARCH

# Required to embed build info into binary.
COPY .git /.git

WORKDIR /workspace

COPY . .

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} make build

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/bin/nimbus .
USER 65532:65532

ENTRYPOINT ["/nimbus"]
