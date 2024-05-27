# SPDX-License-Identifier: Apache-2.0
# Copyright 2023 Authors of Nimbus

# Image URL to use all building/pushing image targets
IMG ?= 5gsec/nimbus-netpol
# Image Tag to use all building/pushing image targets
TAG ?= latest

CONTAINER_TOOL ?= docker
BINARY ?= bin/nimbus-netpol

build:
	@go build -ldflags="-w" -o ${BINARY}  main.go

run: build
	@./${BINARY}

.PHONY: docker-build
docker-build:
	$(CONTAINER_TOOL) build -t ${IMG}:${TAG} --build-arg VERSION=${TAG} -f ./Dockerfile ../../../

.PHONY: docker-push
docker-push:
	$(CONTAINER_TOOL) push ${IMG}:${TAG}

PLATFORMS ?= linux/arm64,linux/amd64
.PHONY: docker-buildx
docker-buildx:
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name project-v3-builder
	$(CONTAINER_TOOL) buildx use project-v3-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --build-arg VERSION=${TAG} --tag ${IMG}:${TAG} -f Dockerfile.cross ../../../ || { $(CONTAINER_TOOL) buildx rm project-v3-builder; rm Dockerfile.cross; exit 1; }
	- $(CONTAINER_TOOL) buildx rm project-v3-builder
	rm Dockerfile.cross
