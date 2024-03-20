// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package util

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	intentv1 "github.com/5GSEC/nimbus/api/v1"
)

// ExtractNpName extracts the actual NimbusPolicy name from a formatted policy
// name.
func ExtractClusterNpName(policyName string) string {
	words := strings.Split(policyName, "-")
	return strings.Join(words[:len(words)-1], "-")
}

// UpdateNpStatus updates the provided NimbusPolicy status with the number and
// names of its descendant policies that were created. Every adapter is
// responsible for updating the status field of the corresponding NimbusPolicy
// with the number and names of successfully created policies by calling this
// API. This provides feedback to users about the translation and deployment of
// their security intent
func UpdateCnpStatus(ctx context.Context, k8sClient client.Client, currPolicyFullName, cnpName string, decrement bool) error {
	// Since multiple adapters may attempt to update the NimbusPolicy status
	// concurrently, potentially leading to conflicts. To ensure data consistency,
	// retry on write failures. On conflict, the update is retried with an
	// exponential backoff strategy. This provides resilience against potential
	// issues while preventing indefinite retries in case of persistent conflicts.
	if retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestCnp := &intentv1.ClusterNimbusPolicy{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: cnpName}, latestCnp); err != nil {
			return nil
		}

		updateCountAndClusterPoliciesName(latestCnp, currPolicyFullName, decrement)
		if err := k8sClient.Status().Update(ctx, latestCnp); err != nil {
			return err
		}

		return nil
	}); retryErr != nil {
		return retryErr
	}
	return nil
}

func updateCountAndClusterPoliciesName(latestCnp *intentv1.ClusterNimbusPolicy, currPolicyFullName string, decrement bool) {
	if !contains(latestCnp.Status.Policies, currPolicyFullName) {
		latestCnp.Status.NumberOfAdapterPolicies++
		latestCnp.Status.Policies = append(latestCnp.Status.Policies, currPolicyFullName)
	}

	if decrement {
		latestCnp.Status.NumberOfAdapterPolicies--
		for idx, existingPolicyName := range latestCnp.Status.Policies {
			if existingPolicyName == currPolicyFullName {
				latestCnp.Status.Policies = append(latestCnp.Status.Policies[:idx], latestCnp.Status.Policies[idx+1:]...)
				return
			}
		}
	}
}
