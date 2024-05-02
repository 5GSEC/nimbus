// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package intentbinder

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1alpha"
)

// ExtractIntents extract the SecurityIntent from the given SecurityIntentBinding
// or ClusterSecurityIntentBinding objects.
func ExtractIntents(ctx context.Context, c client.Client, object client.Object) []v1.SecurityIntent {
	logger := log.FromContext(ctx)
	var intents []v1.SecurityIntent
	var givenIntents []v1.MatchIntent

	switch obj := object.(type) {
	case *v1.SecurityIntentBinding:
		givenIntents = obj.Spec.Intents
	case *v1.ClusterSecurityIntentBinding:
		givenIntents = obj.Spec.Intents
	}

	for _, intent := range givenIntents {
		var si v1.SecurityIntent
		if err := c.Get(ctx, types.NamespacedName{Name: intent.Name}, &si); err != nil && apierrors.IsNotFound(err) {
			logger.V(2).Info("failed to fetch SecurityIntent", "SecurityIntent.Name", intent.Name)
			continue
		}
		intents = append(intents, si)
	}

	return intents
}
