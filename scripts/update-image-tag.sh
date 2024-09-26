#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
# Copyright 2023 Authors of Nimbus

if ! command -v yq >/dev/null; then
  echo "Installing yq..."
  go install github.com/mikefarah/yq/v4@latest
fi

TAG=$1
DEPLOYMENT_ROOT_DIR="deployments"
DIRECTORIES=("${DEPLOYMENT_ROOT_DIR}/nimbus" "${DEPLOYMENT_ROOT_DIR}/nimbus-k8tls" \
  "${DEPLOYMENT_ROOT_DIR}/nimbus-kubearmor" "${DEPLOYMENT_ROOT_DIR}/nimbus-kyverno" "${DEPLOYMENT_ROOT_DIR}/nimbus-netpol")

echo "Updating tag to $TAG"
for directory in "${DIRECTORIES[@]}"; do
   yq -i ".image.tag = \"$TAG\"" "${directory}/values.yaml"
done
