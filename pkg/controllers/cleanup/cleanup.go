// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package cleanup

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	intentv1 "github.com/5GSEC/nimbus/pkg/api/v1"
)

// Cleanup is a function to clean up SecurityIntent resources.
// It removes all policies associated with each SecurityIntent before deleting the SecurityIntent itself.
func Cleanup(ctx context.Context, k8sClient client.Client, logger logr.Logger) error {

	// Logging the start of the cleanup process.
	logger.Info("Performing cleanup")

	var securityIntentBindings intentv1.SecurityIntentBindingList
	if err := k8sClient.List(ctx, &securityIntentBindings); err != nil {
		logger.Error(err, "Unable to list SecurityIntentBinding resources for cleanup")
		return err
	}

	if len(securityIntentBindings.Items) == 0 {
		logger.Info("No SecurityIntentBinding resources found for cleanup")
		return nil
	}

	// Iterating over each SecurityIntent to delete associated policies.
	for _, binding := range securityIntentBindings.Items {
		bindingCopy := binding
		if err := k8sClient.Delete(ctx, &bindingCopy); err != nil {
			logger.Error(err, "Failed to delete SecurityIntentBinding", "Name", bindingCopy.Name)
			continue
		}
	}
	return nil
}
