// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
)

// Controller is a struct that holds a Kubernetes client and a SecurityIntent.
type Controller struct {
	Client                              client.Client
	SecurityIntentWatcher               *SecurityIntent
	SecurityIntentBindingWatcher        *SecurityIntentBinding
	ClusterSecurityIntentBindingWatcher *ClusterSecurityIntentBinding
}

// NewController creates a new instance of Controller.
func NewController(client client.Client) (*Controller, error) {
	if client == nil {
		return nil, fmt.Errorf("Controller: Client is nil")
	}

	siWatcher, err := NewSecurityIntent(client)
	if err != nil {
		return nil, fmt.Errorf("Controller: Error creating SecurityIntent: %v", err)
	}

	sibWatcher, err := NewSecurityIntentBinding(client)
	if err != nil {
		return nil, fmt.Errorf("Controller: Error creating SecurityIntentBinding: %v", err)
	}

	clusterSibWatcher, err := NewClusterSecurityIntentBinding(client)
	if err != nil {
		return nil, fmt.Errorf("Controller: Error creating ClusterSecurityIntentBinding: %v", err)
	}

	return &Controller{
		Client:                              client,
		SecurityIntentWatcher:               siWatcher,
		SecurityIntentBindingWatcher:        sibWatcher,
		ClusterSecurityIntentBindingWatcher: clusterSibWatcher,
	}, nil
}

func (wc *Controller) Reconcile(ctx context.Context, req ctrl.Request) (*v1.SecurityIntent, *v1.SecurityIntentBinding, *v1.ClusterSecurityIntentBinding, error) {
	logger := log.FromContext(ctx)
	if wc == nil {
		return nil, nil, nil, fmt.Errorf("Controller is nil")
	}

	intent, errIntent := wc.SecurityIntentWatcher.Reconcile(ctx, req)
	if errIntent != nil {
		fmt.Println("failed to process SecurityIntent:", errIntent)
		return nil, nil, nil, errIntent
	}

	binding, errBinding := wc.SecurityIntentBindingWatcher.Reconcile(ctx, req)
	if errBinding != nil {
		fmt.Println("failed to process SecurityIntentBinding:", errBinding)
		return nil, nil, nil, errBinding
	}

	clusterSib, err := wc.ClusterSecurityIntentBindingWatcher.Reconcile(ctx, req)
	if err != nil {
		logger.Error(err, "failed to reconcile ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding", req.Name)
		return nil, nil, nil, err
	}

	return intent, binding, clusterSib, nil
}
