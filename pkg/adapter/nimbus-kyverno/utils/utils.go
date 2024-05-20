// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package utils

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetGVK(kind string) string {
	// Map to store the mappings of kinds to their corresponding API versions
	kindToAPIVersion := map[string]string{
		"deployment":            "apps/v1",
		"pod":                   "v1",
		"statefulset":           "apps/v1",
		"daemonset":             "apps/v1",
		"replicaset":            "apps/v1",
	}

	// Convert kind to lowercase to handle case insensitivity
	kind = strings.ToLower(kind)

	// Retrieve API version from the map
	apiVersion, exists := kindToAPIVersion[kind]
	if !exists {
		return "" 
	}

	switch kind {
	case "replicaset":
		kind = "ReplicaSet"
	case "statefulset":
		kind = "StatefulSet"
	case "daemonset" :
		kind = "DaemonSet"
	default:
		kind = Title(kind)
	}

	// Combine API version and kind to form the GroupVersionKind string
	return fmt.Sprintf("%s/%s", apiVersion, Title(kind))
}

func Title(input string) string {
    toTitle := cases.Title(language.Und)

    return toTitle.String(input)
}
