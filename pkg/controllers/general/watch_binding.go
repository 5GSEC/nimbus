// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package general

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	intentv1 "github.com/5GSEC/nimbus/pkg/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// WatcherBinding is a struct that holds a Kubernetes client.
type WatcherBinding struct {
	Client client.Client // Client to interact with Kubernetes resources.
}

// NewWatcherBinding creates a new instance of WatcherBinding.
func NewWatcherBinding(client client.Client) (*WatcherBinding, error) {
	if client == nil {
		// Return an error if the client is not provided.
		return nil, fmt.Errorf("WatcherBinding: Client is nil")
	}

	// Return a new WatcherBinding instance with the provided client.
	return &WatcherBinding{
		Client: client,
	}, nil
}

// Reconcile handles the reconciliation of the SecurityIntentBinding resources.
func (wb *WatcherBinding) Reconcile(ctx context.Context, req ctrl.Request) (*intentv1.SecurityIntentBinding, error) {
	log := log.FromContext(ctx)

	if wb == nil || wb.Client == nil {
		log.Info("WatcherBinding is nil or Client is nil in Reconcile")
		return nil, fmt.Errorf("WatcherBinding or Client is not initialized")
	}

	binding := &intentv1.SecurityIntentBinding{}
	err := wb.Client.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, binding)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("SecurityIntentBinding resource not found. Ignoring since object must be deleted", "Name", req.Name, "Namespace", req.Namespace)
			return nil, nil
		}
		log.Error(err, "Failed to get SecurityIntentBinding", "Name", req.Name, "Namespace", req.Namespace)
		return nil, err
	}
	return binding, nil
}
