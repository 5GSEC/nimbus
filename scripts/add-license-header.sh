#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
# Copyright 2023 Authors of Nimbus

if ! command -v addlicense >/dev/null; then
  echo "Installing addlicense..."
  go install github.com/google/addlicense@latest
fi

GIT_ROOT=$(git rev-parse --show-toplevel)
LICENSE_HEADER=${GIT_ROOT}/scripts/license.header

if [ -z $1 ]; then
  echo "No Argument Supplied, Checking and Fixing all files from project root"
  addlicense -f ${LICENSE_HEADER} -v ${GIT_ROOT}/**/*.sh ${GIT_ROOT}/**/*.go
  echo "Done"
else
  addlicense -f ${LICENSE_HEADER} -v $1
  echo "Done"
fi