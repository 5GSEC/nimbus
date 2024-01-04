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

	v1 "github.com/5GSEC/nimbus/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

type WatcherNimbusPolicy struct {
	Client client.Client
}

func NewWatcherNimbusPolicy(client client.Client) (*WatcherNimbusPolicy, error) {
	if client == nil {
		// Return an error if the client is not provided.
		return nil, fmt.Errorf("WatcherNimbusPolicy: Client is nil")
	}

	return &WatcherNimbusPolicy{
		Client: client,
	}, nil
}

func (wn *WatcherNimbusPolicy) Reconcile(ctx context.Context, req ctrl.Request) (*v1.NimbusPolicy, error) {
	log := log.FromContext(ctx)

	if wn == nil || wn.Client == nil {
		log.Info("WatcherNimbusPolicy is nil or Client is nil in Reconcile")
		return nil, fmt.Errorf("WatcherNimbusPolicy or Client is not initialized")
	}

	nimbusPol := &v1.NimbusPolicy{}
	err := wn.Client.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, nimbusPol)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("NimbusPolicy resource not found. Ignoring since object must be deleted", "Name", req.Name, "Namespace", req.Namespace)
			return nil, nil
		}
		log.Error(err, "Failed to get NimbusPolicy", "Name", req.Name, "Namespace", req.Namespace)
		return nil, err
	}
	return nimbusPol, nil
}
