# SPDX-License-Identifier: Apache-2.0
# Copyright 2023 Authors of Nimbus

# Image URL to use all building/pushing image targets
IMG ?= 5gsec/nimbus-kyverno
# Image Tag to use all building/pushing image targets
TAG ?= latest

CONTAINER_TOOL ?= docker
BINARY ?= bin/nimbus-kyverno

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

.PHONY: build
build: ## Build nimbus-kyverno executable.
	@go build -ldflags="-w" -o ${BINARY}  .

.PHONY: run
run: build ## Run nimbus-kyverno locally.
	@./${BINARY}

.PHONY: docker-build
docker-build: ## Build nimbus-kyverno container image.
	$(CONTAINER_TOOL) build -t ${IMG}:${TAG} --build-arg VERSION=${TAG} -f ./Dockerfile ../../../

.PHONY: docker-push
docker-push: ## Push nimbus-kyverno container image.
	$(CONTAINER_TOOL) push ${IMG}:${TAG}

PLATFORMS ?= linux/arm64,linux/amd64
.PHONY: docker-buildx
docker-buildx: ## Build and push container image for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name project-v3-builder
	$(CONTAINER_TOOL) buildx use project-v3-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --build-arg VERSION=${TAG} --tag ${IMG}:${TAG} -f Dockerfile.cross ../../../ || { $(CONTAINER_TOOL) buildx rm project-v3-builder; rm Dockerfile.cross; exit 1; }
	- $(CONTAINER_TOOL) buildx rm project-v3-builder
	rm Dockerfile.cross
