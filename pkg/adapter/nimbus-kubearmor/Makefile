# SPDX-License-Identifier: Apache-2.0
# Copyright 2023 Authors of Nimbus

# Image URL to use all building/pushing image targets
IMG ?= 5gsec/nimbus-kubearmor
# Image Tag to use all building/pushing image targets
TAG ?= v0.1

CONTAINER_TOOL ?= docker
BINARY ?= bin/nimbus-kubearmor

build:
	@go build -ldflags="-w" -o ${BINARY}  main.go

run: build
	@./${BINARY}

deploy-rbac:
	kubectl apply -f config/service_account.yaml
	kubectl apply -f config/role.yaml
	kubectl apply -f config/role_binding.yaml

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG}:${TAG} -t ${IMG}:latest --build-arg VERSION=${TAG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}:${TAG}
	$(CONTAINER_TOOL) push ${IMG}:latest

PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx:
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name project-v3-builder
	$(CONTAINER_TOOL) buildx use project-v3-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --build-arg VERSION=${TAG} --tag ${IMG}:${TAG} -f Dockerfile.cross . || { $(CONTAINER_TOOL) buildx rm project-v3-builder; rm Dockerfile.cross; exit 1; }
	- $(CONTAINER_TOOL) buildx rm project-v3-builder
	rm Dockerfile.cross
