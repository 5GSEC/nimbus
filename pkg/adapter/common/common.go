// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package common

type Request struct {
	Name      string
	Namespace string
}

type ContextKey string

const (
	K8sClientKey     ContextKey = "k8sClient"
	NamespaceNameKey ContextKey = "K8tlsNamespace"
)
