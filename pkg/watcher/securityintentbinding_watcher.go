// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/apimachinery/pkg/api/errors"

	v1 "github.com/5GSEC/nimbus/api/v1"
)

// SecurityIntentBinding is a struct that holds a Kubernetes client.
type SecurityIntentBinding struct {
	Client client.Client
}

// NewSecurityIntentBinding creates a new instance of SecurityIntentBinding.
func NewSecurityIntentBinding(client client.Client) (*SecurityIntentBinding, error) {
	if client == nil {
		return nil, fmt.Errorf("SecurityIntentBinding: Client is nil")
	}

	// Return a new SecurityIntentBinding instance with the provided client.
	return &SecurityIntentBinding{
		Client: client,
	}, nil
}

// Reconcile handles the reconciliation of the SecurityIntentBinding resources.
func (wb *SecurityIntentBinding) Reconcile(ctx context.Context, req ctrl.Request) (*v1.SecurityIntentBinding, error) {
	logger := log.FromContext(ctx)

	if wb == nil || wb.Client == nil {
		logger.Info("SecurityIntentBinding is nil or Client is nil in Reconcile")
		return nil, fmt.Errorf("SecurityIntentBinding or Client is not initialized")
	}

	binding := &v1.SecurityIntentBinding{}
	err := wb.Client.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, binding)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		logger.Error(err, "failed to get SecurityIntentBinding", "Name", req.Name, "Namespace", req.Namespace)
		return nil, err
	}
	return binding, nil
}
