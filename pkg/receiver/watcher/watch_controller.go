// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"context"
	"fmt"

	v1 "github.com/5GSEC/nimbus/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WatcherController is a struct that holds a Kubernetes client and a WatcherIntent.
type WatcherController struct {
	Client         client.Client
	WatcherIntent  *WatcherIntent
	WatcherBinding *WatcherBinding
}

// NewWatcherController creates a new instance of WatcherController.
func NewWatcherController(client client.Client) (*WatcherController, error) {
	if client == nil {
		return nil, fmt.Errorf("WatcherController: Client is nil")
	}

	watcherIntent, err := NewWatcherIntent(client)
	if err != nil {
		return nil, fmt.Errorf("WatcherController: Error creating WatcherIntent: %v", err)
	}

	watcherBinding, err := NewWatcherBinding(client)
	if err != nil {
		return nil, fmt.Errorf("WatcherController: Error creating WatcherBinding: %v", err)
	}

	return &WatcherController{
		Client:         client,
		WatcherIntent:  watcherIntent,
		WatcherBinding: watcherBinding,
	}, nil
}

func (wc *WatcherController) Reconcile(ctx context.Context, client client.Client, req ctrl.Request) (*v1.SecurityIntent, *v1.SecurityIntentBinding, error) {
	if wc == nil {
		return nil, nil, fmt.Errorf("WatcherController is nil")
	}

	intent, errIntent := wc.WatcherIntent.Reconcile(ctx, req)
	if errIntent != nil {
		fmt.Println("Failed to process SecurityIntent:", errIntent)
		return nil, nil, errIntent
	}

	binding, errBinding := wc.WatcherBinding.Reconcile(ctx, req)
	if errBinding != nil {
		fmt.Println("Failed to process SecurityIntentBinding:", errBinding)
		return intent, nil, errBinding
	}

	return intent, binding, nil
}
