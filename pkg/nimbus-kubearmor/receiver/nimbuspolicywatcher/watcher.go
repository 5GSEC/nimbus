// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package nimbuspolicywatcher

import (
	"context"
	"log"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/5GSEC/nimbus/api/v1"
)

// NimbusPolicyWatcher is a struct that holds a client for interacting with Kubernetes API.
type NimbusPolicyWatcher struct {
	client.Client
}

// NewNimbusPolicyWatcher creates a new instance of NimbusPolicyWatcher.
// It requires a Kubernetes client for operations.
func NewNimbusPolicyWatcher(client client.Client) *NimbusPolicyWatcher {
	return &NimbusPolicyWatcher{Client: client}
}

// WatchNimbusPolicies continuously watches for changes to NimbusPolicy resources across all namespaces.
// It returns a channel through which the NimbusPolicy objects can be received.
func (npw *NimbusPolicyWatcher) WatchNimbusPolicies(ctx context.Context) (<-chan v1.NimbusPolicy, error) {
	policyChan := make(chan v1.NimbusPolicy)
	// NimbusPolicyWatcher 구조체에 추가

	go func() {
		defer close(policyChan)

		for {
			select {
			case <-ctx.Done():
				// Exit the loop if the context is cancelled
				return
			default:
				var nimbusPolicies v1.NimbusPolicyList
				// Attempt to list all NimbusPolicies in all namespaces
				if err := npw.List(ctx, &nimbusPolicies, client.InNamespace("")); err != nil {
					log.Printf("Error listing NimbusPolicies: %v", err)
					// Wait before retrying in case of an error
					time.Sleep(time.Second * 5)
					continue
				}

				// Send each found NimbusPolicy to the channel
				for _, np := range nimbusPolicies.Items {
					policyChan <- np
				}

				// Wait before checking for new changes
				time.Sleep(time.Second * 10)
			}
		}
	}()

	return policyChan, nil
}
