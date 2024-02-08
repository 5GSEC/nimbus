// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package util

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func IsOrphan(ownerRefs []metav1.OwnerReference, ownerKind string) bool {
	return len(ownerRefs) == 0 || ownerRefs[0].Kind != ownerKind
}
