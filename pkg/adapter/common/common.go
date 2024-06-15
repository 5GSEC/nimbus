// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package common

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Request struct {
	Name      string
	Namespace string
}

type PodData struct {
	Name      string
	Namespace string
	Spec      corev1.PodSpec
}

type DeployData struct {
	Name      string
	Namespace string
	Spec      appsv1.DeploymentSpec
}
