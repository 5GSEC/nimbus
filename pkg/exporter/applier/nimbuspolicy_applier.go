// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package applier

import (
	"context"
	"fmt"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NimbusPolicyApplier is responsible for applying NimbusPolicy objects to the Kubernetes cluster.
type NimbusPolicyApplier struct {
	Client client.Client
}

// NewNimbusPolicyApplier creates a new instance of NimbusPolicyApplier.
func NewNimbusPolicyApplier(client client.Client) *NimbusPolicyApplier {
	return &NimbusPolicyApplier{
		Client: client,
	}
}

func (npa *NimbusPolicyApplier) ApplyNimbusPolicy(ctx context.Context, policy *v1.NimbusPolicy) error {
	logger := log.FromContext(ctx)

	// Check if the NimbusPolicy already exists.
	existingPolicy := &v1.NimbusPolicy{}
	err := npa.Client.Get(ctx, client.ObjectKeyFromObject(policy), existingPolicy)

	// Handle NotFound errors the right way
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("Failed to check for existing NimbusPolicy: %v", err)
		}
		// If it's a NotFound error, create a new Nimbus policy.
		logger.Info("Apply NimbusPolicy", "Policy", policy.Name)
		if err := npa.Client.Create(ctx, policy); err != nil {
			return fmt.Errorf("Failed to Apply NimbusPolicy: %v", err)
		}
	} else {
		// If the policy already exists, update it.
		logger.Info("Update NimbusPolicy", "Policy", policy.Name)
		policy.ResourceVersion = existingPolicy.ResourceVersion
		if err := npa.Client.Update(ctx, policy); err != nil {
			return fmt.Errorf("Failed to update NimbusPolicy: %v", err)
		}
	}

	return nil
}
