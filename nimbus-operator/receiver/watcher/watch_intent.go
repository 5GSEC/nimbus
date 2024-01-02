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

	v1 "github.com/5GSEC/nimbus/nimbus-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// WatcherIntent is a struct that holds a Kubernetes client.
type WatcherIntent struct {
	Client client.Client // Client to interact with Kubernetes resources.
}

// NewWatcherIntent creates a new instance of WatcherIntent.
func NewWatcherIntent(client client.Client) (*WatcherIntent, error) {
	if client == nil {
		// Return an error if the client is not provided.
		return nil, fmt.Errorf("WatcherIntent: Client is nil")
	}

	// Return a new WatcherIntent instance with the provided client.
	return &WatcherIntent{
		Client: client,
	}, nil
}

// Reconcile is the method that handles the reconciliation of the Kubernetes resources.
func (wi *WatcherIntent) Reconcile(ctx context.Context, req ctrl.Request) (*v1.SecurityIntent, error) {
	log := log.FromContext(ctx)

	if wi == nil || wi.Client == nil {
		log.Info("WatcherIntent is nil or Client is nil in Reconcile")
		return nil, fmt.Errorf("WatcherIntent or Client is not initialized")
	}

	intent := &v1.SecurityIntent{}
	err := wi.Client.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, intent)

	if err == nil {
		return intent, nil
	} else {
		if errors.IsNotFound(err) {
			log.Info("SecurityIntent resource not found. Ignoring since object must be deleted", "Name", req.Name, "Namespace", req.Namespace)
			return nil, nil
		}
		log.Error(err, "Failed to get SecurityIntent", "Name", req.Name, "Namespace", req.Namespace)
		return nil, err
	}
}
