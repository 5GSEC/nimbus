// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package enforcer

import (
	"context"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/nimbus-kubearmor/core/applier"
	"github.com/5GSEC/nimbus/pkg/nimbus-kubearmor/core/converter"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PolicyEnforcer manages the conversion and enforcement of Nimbus policies.
type PolicyEnforcer struct {
	converter *converter.PolicyConverter
	applier   *applier.Applier
}

// NewPolicyEnforcer creates a new PolicyEnforcer instance.
func NewPolicyEnforcer(client client.Client) *PolicyEnforcer {
	return &PolicyEnforcer{
		converter: converter.NewPolicyConverter(client),
		applier:   applier.NewApplier(client),
	}
}

// ExportAndApplyPolicy converts a NimbusPolicy to a KubeArmorPolicy and applies it.
func (pe *PolicyEnforcer) Enforcer(ctx context.Context, nimbusPolicy v1.NimbusPolicy) error {
	// Convert NimbusPolicy to KubeArmorPolicy
	kubeArmorPolicy, err := pe.converter.Converter(ctx, nimbusPolicy)
	if err != nil {
		return err
	}

	// Apply the converted KubeArmorPolicy
	return pe.applier.ApplyPolicy(ctx, kubeArmorPolicy)
}
