// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package intentbinder

import (
	"context"

	v1 "github.com/5GSEC/nimbus/nimbus-operator/api/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// BindingInfo holds the names of matched SecurityIntent and SecurityIntentBinding.
type BindingInfo struct {
	IntentNames       []string
	IntentNamespaces  []string
	BindingNames      []string
	BindingNamespaces []string
}

// NewBindingInfo creates a new instance of BindingInfo.
func NewBindingInfo(intentNames []string, intentNamespaces []string, bindingNames []string, bindingNamespaces []string) *BindingInfo {
	return &BindingInfo{
		IntentNames:       intentNames,
		IntentNamespaces:  intentNamespaces,
		BindingNames:      bindingNames,
		BindingNamespaces: bindingNamespaces,
	}
}

func MatchAndBindIntents(ctx context.Context, client client.Client, req ctrl.Request, bindings *v1.SecurityIntentBinding) (*BindingInfo, error) {
	log := log.FromContext(ctx)
	log.Info("Starting intent and binding matching")

	// Fetching SecurityIntent objects.
	var intents []*v1.SecurityIntent
	for _, intentRef := range bindings.Spec.Intents {
		intent := &v1.SecurityIntent{}
		if err := client.Get(ctx, types.NamespacedName{Name: intentRef.Name, Namespace: bindings.Namespace}, intent); err != nil {
			log.Error(err, "Failed to get SecurityIntent", "Name", intentRef.Name)
			continue
		}
		intents = append(intents, intent)
	}

	var matchedIntentNames []string
	var matchedIntentNamespaces []string
	var matchedBindingNames []string
	var matchedBindingNamespaces []string

	// Checking match for SecurityIntent and SecurityIntentBinding.
	for _, intent := range intents {
		matchedIntentNames = append(matchedIntentNames, intent.Name)
		matchedIntentNamespaces = append(matchedIntentNamespaces, intent.Namespace)
	}

	// Adding names and namespaces of SecurityIntentBinding.
	matchedBindingNames = append(matchedBindingNames, bindings.Name)
	matchedBindingNamespaces = append(matchedBindingNamespaces, bindings.Namespace)

	log.Info("Matching completed")
	log.Info("Matching completed", "Matched Intent Names", matchedIntentNames, "Matched Binding Names", matchedBindingNames)
	return NewBindingInfo(matchedIntentNames, matchedIntentNamespaces, matchedBindingNames, matchedBindingNamespaces), nil
}
