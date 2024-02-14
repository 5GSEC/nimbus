// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package policybuilder

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/processor/intentbinder"
)

// BuildNimbusPolicy generates a NimbusPolicy based on SecurityIntent and SecurityIntentBinding.
func BuildNimbusPolicy(ctx context.Context, client client.Client, scheme *runtime.Scheme, bindingInfo *intentbinder.BindingInfo) (*v1.NimbusPolicy, error) {
	logger := log.FromContext(ctx)
	logger.Info("Building NimbusPolicy")

	var nimbusRulesList []v1.NimbusRules

	// Iterate over intent names to build rules.
	for _, intentName := range bindingInfo.IntentNames {
		intent, err := FetchIntentByName(ctx, client, intentName)
		if err != nil {
			return nil, err
		}

		// Checks if arrays in bindingInfo are empty.
		if len(bindingInfo.IntentNames) == 0 || len(bindingInfo.BindingNames) == 0 {
			fmt.Println("No intents or bindings to process")
			return nil, fmt.Errorf("no intents or bindings to process")
		}

		// Constructs a rule from the intent parameters.
		rule := v1.Rule{
			RuleAction: intent.Spec.Intent.Action,
			Params:     map[string][]string{},
		}

		for key, val := range intent.Spec.Intent.Params {
			rule.Params[key] = val
		}

		nimbusRule := v1.NimbusRules{
			ID:          intent.Spec.Intent.ID,
			Type:        "", // Set Type if necessary
			Description: intent.Spec.Intent.Description,
			Rule:        rule,
		}
		nimbusRulesList = append(nimbusRulesList, nimbusRule)
	}

	// Fetches the binding to extract selector.
	bindingName := bindingInfo.BindingNames[0]
	bindingNamespace := bindingInfo.BindingNamespaces[0]
	binding, err := fetchBinding(ctx, client, bindingName, bindingNamespace)
	if err != nil {
		return nil, err
	}

	// Extracts match labels from the binding selector.
	matchLabels, err := extractSelector(ctx, client, bindingNamespace, binding.Spec.Selector)
	if err != nil {
		return nil, err
	}

	if len(matchLabels) == 0 {
		logger.Error(err, "No labels matched the CEL expressions, aborting NimbusPolicy creation due to missing keys in labels")
		return nil, nil
	}

	// Creates a NimbusPolicy.
	nimbusPolicy := &v1.NimbusPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      binding.Name,
			Namespace: binding.Namespace,
		},
		Spec: v1.NimbusPolicySpec{
			Selector: v1.NimbusSelector{
				MatchLabels: matchLabels,
			},
			NimbusRules: nimbusRulesList,
		},
		Status: v1.NimbusPolicyStatus{
			Status: "Pending",
		},
	}

	if err = ctrl.SetControllerReference(binding, nimbusPolicy, scheme); err != nil {
		logger.Error(err, "failed to set OwnerReference")
		return nil, err
	}

	logger.Info("NimbusPolicy built successfully", "Policy", nimbusPolicy)
	return nimbusPolicy, nil
}

// fetchBinding fetches a SecurityIntentBinding by its name and namespace.
func fetchBinding(ctx context.Context, client client.Client, name string, namespace string) (*v1.SecurityIntentBinding, error) {
	logger := log.FromContext(ctx)
	var binding v1.SecurityIntentBinding
	if err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &binding); err != nil {
		logger.Error(err, "Failed to get SecurityIntentBinding")
		return nil, err
	}
	return &binding, nil
}

// extractSelector extracts match labels from a Selector.
func extractSelector(ctx context.Context, k8sClient client.Client, namespace string, selector v1.Selector) (map[string]string, error) {
	matchLabels := make(map[string]string) // Initialize map for match labels.

	// Process CEL expressions.
	if len(selector.CEL) > 0 {
		celExpressions := selector.CEL
		celMatchLabels, err := ProcessCEL(ctx, k8sClient, namespace, celExpressions)
		if err != nil {
			return nil, fmt.Errorf("error processing CEL: %v", err)
		}
		for k, v := range celMatchLabels {
			matchLabels[k] = v
		}
	}

	// Process Any/All fields.
	if len(selector.Any) > 0 || len(selector.All) > 0 {
		matchLabelsFromAnyAll, err := ProcessMatchLabels(selector.Any, selector.All)
		if err != nil {
			return nil, fmt.Errorf("error processing Any/All match labels: %v", err)
		}
		for key, value := range matchLabelsFromAnyAll {
			matchLabels[key] = value
		}
	}

	return matchLabels, nil
}
