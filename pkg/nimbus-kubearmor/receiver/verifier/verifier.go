// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package verifier

import (
	"strings"

	v1 "github.com/5GSEC/nimbus/api/v1"
)

// HandlePolicy checks if the given NimbusPolicy contains any rules that start with "sys".
// It iterates through the NimbusRules in the policy and returns true if any rule's Id starts with "sys".
// This function is used to identify policies that should be processed by this adapter.
func HandlePolicy(policy v1.NimbusPolicy) bool {
	for _, rule := range policy.Spec.NimbusRules {
		if strings.HasPrefix(rule.Id, "sys") {
			// If any rule's Id starts with "sys", return true
			return true
		}
	}
	// Return false if no such rules are found
	return false
}
