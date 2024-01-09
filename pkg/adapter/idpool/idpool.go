// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

// Package idpool manages a pool of IDs for use by adapters.
package idpool

import (
	"strings"
)

const (
	SwDeploymentTools     = "swDeploymentTools"
	UnAuthorizedNEFAccess = "unAuthorizedNEFAccess"
	NFServiceDiscovery    = "nfServiceDiscovery"
	DNSManipulation       = "dnsManipulation"
	NetPortExec           = "netPortExec"
	SysPathExec           = "sysPathExec"
)

// KaIds are IDs supported by KubeArmor security engine.
var KaIds = []string{
	SwDeploymentTools,
}

// IdSupportedBy determines whether a given ID is supported by a security engine.
func IdSupportedBy(id, securityEngine string) bool {
	switch strings.ToLower(securityEngine) {
	case "kubearmor":
		return in(id, KaIds)
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
