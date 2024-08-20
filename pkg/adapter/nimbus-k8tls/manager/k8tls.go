// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package manager

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func k8tlsEnvExist(ctx context.Context, k8sClient client.Client) bool {
	logger := log.FromContext(ctx)

	ns := &corev1.Namespace{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Name: K8tlsNamespace}, ns); err != nil {
		logger.Error(err, "'k8tls' namespace not found")
		return false
	}

	sa := &corev1.ServiceAccount{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Name: k8tls, Namespace: K8tlsNamespace}, sa); err != nil {
		logger.Error(err, "'k8tls' serviceaccount not found")
		return false
	}

	// If the required ClusterRole and ClusterRoleBinding resources don't exist, the
	// job itself will describe/log that error.
	return true
}
