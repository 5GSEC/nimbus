// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package util

import (
	"strings"
)

func ExtractNpName(kspName string) string {
	words := strings.Split(kspName, "-")
	return strings.Join(words[:len(words)-1], "-")
}
