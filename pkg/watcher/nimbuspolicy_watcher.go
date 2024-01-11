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

type NimbusPolicy struct {
	Client client.Client
}

func NewNimbusPolicy(client client.Client) (*NimbusPolicy, error) {
	if client == nil {
		return nil, fmt.Errorf("NimbusPolicy: Client is nil")
	}

	return &NimbusPolicy{
		Client: client,
	}, nil
}

func (wn *NimbusPolicy) Reconcile(ctx context.Context, req ctrl.Request) (*v1.NimbusPolicy, error) {
	logger := log.FromContext(ctx)

	if wn == nil || wn.Client == nil {
		logger.Info("NimbusPolicy is nil or Client is nil in Reconcile")
		return nil, fmt.Errorf("NimbusPolicy or Client is not initialized")
	}

	nimbusPol := &v1.NimbusPolicy{}
	err := wn.Client.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, nimbusPol)

	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("NimbusPolicy resource not found. Ignoring since object must be deleted", "Name", req.Name, "Namespace", req.Namespace)
			return nil, nil
		}
		logger.Error(err, "failed to get NimbusPolicy", "Name", req.Name, "Namespace", req.Namespace)
		return nil, err
	}
	return nimbusPol, nil
}
