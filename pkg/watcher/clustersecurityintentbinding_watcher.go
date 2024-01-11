// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
)

type ClusterSecurityIntentBinding struct {
	Client client.Client
}

func NewClusterSecurityIntentBinding(client client.Client) (*ClusterSecurityIntentBinding, error) {
	if client == nil {
		return nil, fmt.Errorf("ClusterSecurityIntentBinding: Client is nil")
	}

	return &ClusterSecurityIntentBinding{
		Client: client,
	}, nil
}

func (wcb *ClusterSecurityIntentBinding) Reconcile(ctx context.Context, req ctrl.Request) (*v1.ClusterSecurityIntentBinding, error) {
	logger := log.FromContext(ctx)
	if wcb == nil || wcb.Client == nil {
		logger.Info("ClusterSecurityIntentBinding is nil or Client is nil in Reconcile")
		return nil, fmt.Errorf("ClusterSecurityIntentBinding or Client is not initialized")
	}

	clusterSib := &v1.ClusterSecurityIntentBinding{}
	err := wcb.Client.Get(ctx, types.NamespacedName{Name: req.Name}, clusterSib)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		logger.Error(err, "failed to get ClusterSecurityIntentBinding", "Name", req.Name)
		return nil, err
	}
	return clusterSib, nil
}
