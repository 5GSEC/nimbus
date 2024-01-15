// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

// Package idpool manages a pool of IDs for use by KubeArmor.
package idpool

const (
	SwDeploymentTools         = "swDeploymentTools"
	UnAuthorizedSaTokenAccess = "unAuthorizedSaTokenAccess"
	UnAuthorizedNEFAccess     = "unAuthorizedNEFAccess"
	NFServiceDiscovery        = "nfServiceDiscovery"
	DNSManipulation           = "dnsManipulation"
	NetPortExec               = "netPortExec"
	SysPathExec               = "sysPathExec"
)

// KaIds are IDs supported by KubeArmor.
var KaIds = []string{
	SwDeploymentTools, UnAuthorizedSaTokenAccess,
}

// IsIdSupported determines whether a given ID is supported by KubeArmor.
func IsIdSupported(id string) bool {
	for _, currId := range KaIds {
		if currId == id {
			return true
		}
	}
	return false
}
