// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package intentbinder

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
)

// BindingInfo holds the names of matched SecurityIntent and SecurityIntentBinding.
type BindingInfo struct {
	IntentNames       []string
	BindingNames      []string
	BindingNamespaces []string
}

// NewBindingInfo creates a new instance of BindingInfo.
func NewBindingInfo(intentNames []string, bindingNames []string, bindingNamespaces []string) *BindingInfo {
	return &BindingInfo{
		IntentNames:       intentNames,
		BindingNames:      bindingNames,
		BindingNamespaces: bindingNamespaces,
	}
}

func MatchAndBindIntents(ctx context.Context, client client.Client, bindings *v1.SecurityIntentBinding) *BindingInfo {
	logger := log.FromContext(ctx)
	logger.Info("SecurityIntent and SecurityIntentBinding matching started")

	var matchedIntents []string
	var matchedBindings []string
	var matchedBindingNamespaces []string

	for _, intentRef := range bindings.Spec.Intents {
		var intent v1.SecurityIntent
		if err := client.Get(ctx, types.NamespacedName{Name: intentRef.Name, Namespace: bindings.Namespace}, &intent); err != nil {
			logger.Info("failed to get SecurityIntent", "SecurityIntent.Name", intentRef.Name)
			continue
		}
		matchedIntents = append(matchedIntents, intent.Name)
	}

	// Adding names and namespaces of SecurityIntentBinding.
	matchedBindings = append(matchedBindings, bindings.Name)
	matchedBindingNamespaces = append(matchedBindingNamespaces, bindings.Namespace)

	logger.Info("Matching completed", "Matched SecurityIntents", matchedIntents, "Matched SecurityIntentsBindings", matchedBindings)
	return NewBindingInfo(matchedIntents, matchedBindings, matchedBindingNamespaces)
}

func MatchAndBindIntentsGlobal(ctx context.Context, client client.Client, clusterBinding *v1.ClusterSecurityIntentBinding) *BindingInfo {
	logger := log.FromContext(ctx)
	logger.Info("SecurityIntent and ClusterSecurityIntentBinding matching started")

	var matchedIntents []string
	for _, intentRef := range clusterBinding.Spec.Intents {
		var intent v1.SecurityIntent
		if err := client.Get(ctx, types.NamespacedName{Name: intentRef.Name}, &intent); err != nil {
			logger.Info("failed to get SecurityIntent", "SecurityIntent.Name", intentRef.Name)
			continue
		}
		matchedIntents = append(matchedIntents, intent.Name)
	}

	var matchedClusterBindings []string
	matchedClusterBindings = append(matchedClusterBindings, clusterBinding.Name)

	logger.Info("Matching completed", "Matched SecurityIntents", matchedIntents, "Matched ClusterSecurityIntentBindings", matchedClusterBindings)
	return NewBindingInfo(matchedIntents, matchedClusterBindings, nil)
}
