// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package nimbuspolicybuilder

import (
	"context"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/processor/intentbinder"
)

// BuildNimbusPolicy generates a NimbusPolicy based on SecurityIntent and SecurityIntentBinding.
func BuildNimbusPolicy(ctx context.Context, client client.Client, bindingInfo *intentbinder.BindingInfo) (*v1.NimbusPolicy, error) {
	log := log.FromContext(ctx)
	log.Info("Starting NimbusPolicy building")

	var nimbusRulesList []v1.NimbusRules

	// Iterate over intent names to build rules.
	for _, intentName := range bindingInfo.IntentNames {
		intent, err := fetchIntentByName(ctx, client, intentName)
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
			Mode:       intent.Spec.Intent.Mode,
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
	binding, err := fetchBindingByName(ctx, client, bindingName, bindingNamespace)
	if err != nil {
		return nil, err
	}

	// Extracts match labels from the binding selector.
	matchLabels, err := extractSelector(binding.Spec.Selector)
	if err != nil {
		return nil, err
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

	log.Info("NimbusPolicy built successfully", "Policy", nimbusPolicy)
	return nimbusPolicy, nil
}

// fetchIntentByName fetches a SecurityIntent by its name and namespace.
func fetchIntentByName(ctx context.Context, client client.Client, name string) (*v1.SecurityIntent, error) {
	log := log.FromContext(ctx)

	var intent v1.SecurityIntent
	if err := client.Get(ctx, types.NamespacedName{Name: name}, &intent); err != nil {
		log.Error(err, "Failed to get SecurityIntent")
		return nil, err
	}
	return &intent, nil
}

// fetchBindingByName fetches a SecurityIntentBinding by its name and namespace.
func fetchBindingByName(ctx context.Context, client client.Client, name string, namespace string) (*v1.SecurityIntentBinding, error) {
	log := log.FromContext(ctx)
	var binding v1.SecurityIntentBinding
	if err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &binding); err != nil {
		log.Error(err, "Failed to get SecurityIntentBinding")
		return nil, err
	}
	return &binding, nil
}

// extractSelector extracts match labels from a Selector.
func extractSelector(selector v1.Selector) (map[string]string, error) {
	matchLabels := make(map[string]string) // Initialize map for match labels.

	// Process CEL expressions.
	if len(selector.CEL) > 0 {
		celMatchLabels, err := ProcessCEL(selector.CEL)
		if err != nil {
			return nil, fmt.Errorf("Error processing CEL: %v", err)
		}
		for k, v := range celMatchLabels {
			matchLabels[k] = v
		}
	}

	// Process Any/All fields.
	if len(selector.Any) > 0 || len(selector.All) > 0 {
		matchLabelsFromAnyAll, err := ProcessMatchLabels(selector.Any, selector.All)
		if err != nil {
			return nil, fmt.Errorf("Error processing Any/All match labels: %v", err)
		}
		for key, value := range matchLabelsFromAnyAll {
			matchLabels[key] = value
		}
	}

	return matchLabels, nil
}

// ProcessCEL processes CEL expressions to generate matchLabels.
func ProcessCEL(expressions []string) (map[string]string, error) {
	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("label", decls.NewMapType(decls.String, decls.String)), // Define label as a map of string to string.
		),
	)
	if err != nil {
		return nil, fmt.Errorf("Error creating CEL environment: %v", err)
	}

	matchLabels := make(map[string]string)

	for _, expr := range expressions {
		ast, issues := env.Compile(expr)
		if issues != nil && issues.Err() != nil {
			return nil, fmt.Errorf("Error compiling CEL expression: %v", issues.Err())
		}

		prg, err := env.Program(ast)
		if err != nil {
			return nil, fmt.Errorf("Error creating CEL program: %v", err)
		}

		out, _, err := prg.Eval(map[string]interface{}{
			"label": map[string]interface{}{},
		})
		if err != nil {
			return nil, fmt.Errorf("Error evaluating CEL expression: %v", err)
		}

		// Handle the output of the CEL expression.
		if outValue, ok := out.Value().(map[string]interface{}); ok {
			for k, v := range outValue {
				if val, ok := v.(string); ok {
					matchLabels[k] = val
				}
			}
		}
	}
	return matchLabels, nil
}

// ProcessMatchLabels processes any/all fields to generate matchLabels.
func ProcessMatchLabels(any, all []v1.ResourceFilter) (map[string]string, error) {
	matchLabels := make(map[string]string)

	// Process logic for Any field.
	for _, filter := range any {
		for key, value := range filter.Resources.MatchLabels {
			matchLabels[key] = value
		}
	}

	// Process logic for All field.
	for _, filter := range all {
		for key, value := range filter.Resources.MatchLabels {
			matchLabels[key] = value
		}
	}

	return matchLabels, nil
}

func BuildClusterNimbusPolicy(ctx context.Context, client client.Client, clusterBindingInfo *intentbinder.BindingInfo) (*v1.ClusterNimbusPolicy, error) {
	logger := log.FromContext(ctx)
	logger.Info("Building ClusterNimbusPolicy")

	var nimbusRules []v1.NimbusRules
	for _, intentName := range clusterBindingInfo.IntentNames {
		intent, err := fetchIntentByName(ctx, client, intentName)
		if err != nil {
			return nil, err
		}

		if len(clusterBindingInfo.IntentNames) == 0 || len(clusterBindingInfo.BindingNames) == 0 {
			logger.Info("No SecurityIntents or SecurityIntentsBindings to process")
			return nil, fmt.Errorf("no SecurityIntents or SecurityIntentsBindings to process")
		}

		rule := v1.Rule{
			RuleAction: intent.Spec.Intent.Action,
			Mode:       intent.Spec.Intent.Mode,
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
		nimbusRules = append(nimbusRules, nimbusRule)
	}

	binding, err := fetchClusterBindingByName(ctx, client, clusterBindingInfo.BindingNames[0])
	if err != nil {
		return nil, err
	}

	clusterBindingSelector := extractClusterBindingSelector(binding.Spec.Selector)

	clusterNimbusPolicy := &v1.ClusterNimbusPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: binding.Name,
		},
		Spec: v1.ClusterNimbusPolicySpec{
			Selector:    clusterBindingSelector,
			NimbusRules: nimbusRules,
		},
		Status: v1.ClusterNimbusPolicyStatus{
			Status: "Pending",
		},
	}

	logger.Info("ClusterNimbusPolicy built successfully", "ClusterNimbusPolicy", clusterNimbusPolicy)
	return clusterNimbusPolicy, nil
}

func extractClusterBindingSelector(cwSelector v1.CwSelector) v1.CwSelector {
	var clusterBindingSelector v1.CwSelector
	for _, resource := range cwSelector.Resources {
		var cwresource v1.CwResource
		cwresource.Kind = resource.Kind
		cwresource.Name = resource.Name
		cwresource.Namespace = resource.Namespace
		cwresource.MatchLabels = resource.MatchLabels
		clusterBindingSelector.Resources = append(clusterBindingSelector.Resources, cwresource)
	}
	return clusterBindingSelector
}

func fetchClusterBindingByName(ctx context.Context, client client.Client, clusterBindingName string) (v1.ClusterSecurityIntentBinding, error) {
	logger := log.FromContext(ctx)
	var clusterBinding v1.ClusterSecurityIntentBinding
	if err := client.Get(ctx, types.NamespacedName{Name: clusterBindingName}, &clusterBinding); err != nil {
		logger.Error(err, "failed to get ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding", clusterBindingName)
		return v1.ClusterSecurityIntentBinding{}, err
	}
	return clusterBinding, nil
}
