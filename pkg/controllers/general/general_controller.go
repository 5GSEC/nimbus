// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package general

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GeneralController is a struct that holds a Kubernetes client and a WatcherIntent.
type GeneralController struct {
	Client         client.Client   // Client is used to interact with the Kubernetes API.
	WatcherIntent  *WatcherIntent  // WatcherIntent is a custom struct to manage specific operations.
	WatcherBinding *WatcherBinding // WatcherBinding is a custom struct to manage SecurityIntentBinding operations.
}

// NewGeneralController creates a new instance of GeneralController.
func NewGeneralController(client client.Client) (*GeneralController, error) {
	if client == nil {
		// If the client is not provided, return an error.
		return nil, fmt.Errorf("GeneralController: Client is nil")
	}

	// Create a new WatcherIntent.
	watcherIntent, err := NewWatcherIntent(client)
	if err != nil {
		// If there is an error in creating WatcherIntent, return an error.
		return nil, fmt.Errorf("GeneralController: Error creating WatcherIntent: %v", err)
	}

	// Create a new WatcherBinding.
	watcherBinding, err := NewWatcherBinding(client)
	if err != nil {
		// If there is an error in creating WatcherBinding, return an error.
		return nil, fmt.Errorf("GeneralController: Error creating WatcherBinding: %v", err)
	}

	// Return a new GeneralController instance with initialized fields.
	return &GeneralController{
		Client:         client,
		WatcherIntent:  watcherIntent,
		WatcherBinding: watcherBinding,
	}, nil
}

func (gc *GeneralController) Reconcile(ctx context.Context, req ctrl.Request) (*BindingInfo, error) {
	if gc == nil {
		return nil, fmt.Errorf("GeneralController is nil")
	}

	if gc.WatcherIntent == nil {
		return nil, fmt.Errorf("WatcherIntent is nil")
	}

	intent, err := gc.WatcherIntent.Reconcile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Error in WatcherIntent.Reconcile: %v", err)
	}

	if intent != nil {
		return nil, nil
	}

	if gc.WatcherBinding == nil {
		return nil, fmt.Errorf("WatcherBinding is nil")
	}

	binding, err := gc.WatcherBinding.Reconcile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Error in WatcherBinding.Reconcile: %v", err)
	}

	if binding != nil {
		bindingInfo, err := MatchIntentAndBinding(ctx, gc.Client, binding)
		if err != nil {
			return nil, fmt.Errorf("Error in MatchIntentAndBinding: %v", err)
		}
		return bindingInfo, nil
	}

	return nil, nil
}
