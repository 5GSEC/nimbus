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

// SecurityIntent is a struct that holds a Kubernetes client.
type SecurityIntent struct {
	Client client.Client // Client to interact with Kubernetes resources.
}

// NewSecurityIntent creates a new instance of SecurityIntent.
func NewSecurityIntent(client client.Client) (*SecurityIntent, error) {
	if client == nil {
		return nil, fmt.Errorf("SecurityIntent: Client is nil")
	}

	return &SecurityIntent{
		Client: client,
	}, nil
}

// Reconcile is the method that handles the reconciliation of the Kubernetes resources.
func (wi *SecurityIntent) Reconcile(ctx context.Context, req ctrl.Request) (*v1.SecurityIntent, error) {
	logger := log.FromContext(ctx)

	if wi == nil || wi.Client == nil {
		logger.Info("SecurityIntent is nil or Client is nil in Reconcile")
		return nil, fmt.Errorf("SecurityIntent or Client is not initialized")
	}

	intent := &v1.SecurityIntent{}
	err := wi.Client.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, intent)

	if err == nil {
		return intent, nil
	} else {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		logger.Error(err, "failed to get SecurityIntent", "Name", req.Name, "Namespace", req.Namespace)
		return nil, err
	}
}
