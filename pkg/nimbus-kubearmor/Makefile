# SPDX-License-Identifier: Apache-2.0
# Copyright 2023 Authors of Nimbus

.PHONY: build run deploy-rbac

BINARY_NAME=nimbus-kubearmor

build:
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME} main.go

run:
	./nimbus-kubearmor

deploy-rbac:
	kubectl apply -f config/service_account.yaml
	kubectl apply -f config/role.yaml
	kubectl apply -f config/role_binding.yaml
