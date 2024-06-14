// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package util

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func IsOrphan(ownerRefs []metav1.OwnerReference, ownerKind ...string) bool {
	if len(ownerRefs) == 0 {
		return true
	}
	for _, oKind := range ownerKind {
		if ownerRefs[0].Kind == oKind {
			return false
		}
	}
	return true
}
