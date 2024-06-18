// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

// Package idpool manages a pool of IDs for use by adapters.
package idpool

import (
	"strings"
)

const (
	SwDeploymentTools         = "swDeploymentTools"
	UnAuthorizedSaTokenAccess = "unAuthorizedSaTokenAccess"
	DNSManipulation           = "dnsManipulation"
	EscapeToHost              = "escapeToHost"
	DisallowChRoot            = "disallowChRoot"
	DisallowCapabilities      = "disallowCapabilities"
	ExploitPFA                = "PreventExecutionFromTempOrLogsFolders"
	EnsureTLS                 = "ensureTLS"
)

// KaIds are IDs supported by KubeArmor.
var KaIds = []string{
	SwDeploymentTools, UnAuthorizedSaTokenAccess, DNSManipulation, EscapeToHost, ExploitPFA,
}

// list of policies which satisfies the given ID by Kubearmor
var KaIDPolicies = map[string][]string{
	EscapeToHost: {
		DisallowChRoot,
		DisallowCapabilities,
		SwDeploymentTools,
	},
}

// NetPolIDs are IDs supported by Network Policy adapter.
var NetPolIDs = []string{
	DNSManipulation,
}

// KyvIds are IDs supported by Kyverno.
var KyvIds = []string{
	EscapeToHost,
}

// k8tlsIds are IDs supported by k8tls.
var k8tlsIds = []string{
	EnsureTLS,
}

// IsIdSupportedBy determines whether a given ID is supported by a security engine.
func IsIdSupportedBy(id, securityEngine string) bool {
	switch strings.ToLower(securityEngine) {
	case "kubearmor":
		return in(id, KaIds)
	case "netpol":
		return in(id, NetPolIDs)
	case "kyverno":
		return in(id, KyvIds)
	case "k8tls":
		return in(id, k8tlsIds)
	default:
		return false
	}
}

func in(id string, securityEngineIds []string) bool {
	for _, currId := range securityEngineIds {
		if currId == id {
			return true
		}
	}
	return false
}
