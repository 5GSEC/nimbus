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
func ExtractNpName(policyName string) string {
	words := strings.Split(policyName, "-")
	return strings.Join(words[:len(words)-1], "-")
}

// UpdateNpStatus updates the provided NimbusPolicy status with the number and
// names of its descendant policies that were created.
func UpdateNpStatus(ctx context.Context, k8sClient client.Client, currPolicyFullName, npName, namespace string) error {
	// Since multiple adapters may attempt to update the NimbusPolicy status
	// concurrently, potentially leading to conflicts. To ensure data consistency,
	// retry on write failures. On conflict, the update is retried with an
	// exponential backoff strategy. This provides resilience against potential
	// issues while preventing indefinite retries in case of persistent conflicts.
	if retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestNp := &intentv1.NimbusPolicy{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: npName, Namespace: namespace}, latestNp); err != nil {
			return nil
		}

		updateCountAndPoliciesName(latestNp, currPolicyFullName)
		if err := k8sClient.Status().Update(ctx, latestNp); err != nil {
			return err
		}

		return nil
	}); retryErr != nil {
		return retryErr
	}
	return nil
}

func updateCountAndPoliciesName(latestNp *intentv1.NimbusPolicy, currPolicyFullName string) {
	if !contains(latestNp.Status.Policies, currPolicyFullName) {
		latestNp.Status.NumberOfAdapterPolicies++
		latestNp.Status.Policies = append(latestNp.Status.Policies, currPolicyFullName)
	}
}

func contains(existingPolicies []string, policy string) bool {
	for _, existingPolicy := range existingPolicies {
		if existingPolicy == policy {
			return true
		}
	}
	return false
}
