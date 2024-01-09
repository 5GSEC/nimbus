# SPDX-License-Identifier: Apache-2.0
# Copyright 2023 Authors of Nimbus

.PHONY: build run deploy-rbac

build:
	go build -o nimbus-kubearmor

run:
	./nimbus-kubearmor

deploy-rbac:
	kubectl apply -f config/service_account.yaml
	kubectl apply -f config/role.yaml
	kubectl apply -f config/role_binding.yaml
