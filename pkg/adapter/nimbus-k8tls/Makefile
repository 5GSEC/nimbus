# SPDX-License-Identifier: Apache-2.0
# Copyright 2023 Authors of Nimbus

# Image URL to use all building/pushing image targets
IMG ?= 5gsec/nimbus-k8tls
# Image Tag to use all building/pushing image targets
TAG ?= latest

CONTAINER_TOOL ?= docker
BINARY ?= bin/nimbus-k8tls
CONTROLLER_TOOLS_VERSION ?= v0.14.0
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

.PHONY: build
build: ## Build nimbus-k8tls executable.
	@go build -ldflags="-s" -o ${BINARY} .

.PHONY: run
run: build ## Run nimbus-k8tls.
	@./${BINARY}

.PHONY: docker-build
docker-build: ## Build nimbus-k8tls container image.
	$(CONTAINER_TOOL) build -t ${IMG}:${TAG} --build-arg VERSION=${TAG} -f ./Dockerfile ../../../

.PHONY: docker-push
docker-push: ## Push nimbus-k8tls container image.
	$(CONTAINER_TOOL) push ${IMG}:${TAG}

PLATFORMS ?= linux/arm64,linux/amd64
.PHONY: docker-buildx
docker-buildx: ## Build and push container image for cross-platform support
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name project-v3-builder
	$(CONTAINER_TOOL) buildx use project-v3-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --build-arg VERSION=${TAG} --tag ${IMG}:${TAG} -f Dockerfile.cross ../../../ || { $(CONTAINER_TOOL) buildx rm project-v3-builder; rm Dockerfile.cross; exit 1; }
	- $(CONTAINER_TOOL) buildx rm project-v3-builder
	rm Dockerfile.cross

.PHONY: manifests
manifests: controller-gen ## Generate ClusterRole.
	$(CONTROLLER_GEN) rbac:roleName=nimbus-k8tls-clusterrole paths="./..." output:dir=../../../deployments/nimbus-k8tls/templates/

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
