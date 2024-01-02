// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package nimbuspolicybuilder

import (
	"context"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	ctrl "sigs.k8s.io/controller-runtime"

	v1 "github.com/5GSEC/nimbus/nimbus-operator/api/v1"
	intentbinder "github.com/5GSEC/nimbus/nimbus-operator/processor/intentbinder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// BuildNimbusPolicy generates a NimbusPolicy based on SecurityIntent and SecurityIntentBinding.
func BuildNimbusPolicy(ctx context.Context, client client.Client, req ctrl.Request, bindingInfo *intentbinder.BindingInfo) (*v1.NimbusPolicy, error) {
	log := log.FromContext(ctx)
	log.Info("Starting NimbusPolicy building")

	// Validates bindingInfo.
	if bindingInfo == nil || len(bindingInfo.IntentNames) == 0 || len(bindingInfo.IntentNamespaces) == 0 ||
		len(bindingInfo.BindingNames) == 0 || len(bindingInfo.BindingNamespaces) == 0 {
		return nil, fmt.Errorf("invalid bindingInfo: one or more arrays are empty")
	}

	var nimbusRulesList []v1.NimbusRules

	// Iterate over intent names to build rules.
	for i, intentName := range bindingInfo.IntentNames {
		// Checks for array length consistency.
		if i >= len(bindingInfo.IntentNamespaces) || i >= len(bindingInfo.BindingNames) ||
			i >= len(bindingInfo.BindingNamespaces) {
			return nil, fmt.Errorf("index out of range in bindingInfo arrays")
		}

		intentNamespace := bindingInfo.IntentNamespaces[i]
		intent, err := fetchIntentByName(ctx, client, intentName, intentNamespace)
		if err != nil {
			return nil, err
		}

		// Checks if arrays in bindingInfo are empty.
		if len(bindingInfo.IntentNames) == 0 || len(bindingInfo.BindingNames) == 0 {
			fmt.Println("No intents or bindings to process")
			return nil, fmt.Errorf("no intents or bindings to process")
		}

		var rules []v1.Rule

		// Constructs a rule from the intent parameters.
		rule := v1.Rule{
			RuleAction:        intent.Spec.Intent.Action,
			MatchProtocols:    []v1.MatchProtocol{},
			MatchPaths:        []v1.MatchPath{},
			MatchDirectories:  []v1.MatchDirectory{},
			MatchPatterns:     []v1.MatchPattern{},
			MatchCapabilities: []v1.MatchCapability{},
			MatchSyscalls:     []v1.MatchSyscall{},
			FromCIDRSet:       []v1.CIDRSet{},
			ToPorts:           []v1.ToPort{},
		}

		for _, param := range intent.Spec.Intent.Params {
			processSecurityIntentParams(&rule, param)
		}
		rules = append(rules, rule)

		nimbusRule := v1.NimbusRules{
			Id:          intent.Spec.Intent.Id,
			Type:        "", // Set Type if necessary
			Description: intent.Spec.Intent.Description,
			Rule:        rules,
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
			PolicyStatus: "Pending",
		},
	}

	log.Info("NimbusPolicy built successfully", "Policy", nimbusPolicy)
	return nimbusPolicy, nil
}

// fetchIntentByName fetches a SecurityIntent by its name and namespace.
func fetchIntentByName(ctx context.Context, client client.Client, name string, namespace string) (*v1.SecurityIntent, error) {
	log := log.FromContext(ctx)

	var intent v1.SecurityIntent
	if err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &intent); err != nil {
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

// processSecurityIntentParams processes the parameters of a SecurityIntent.
func processSecurityIntentParams(rule *v1.Rule, param v1.SecurityIntentParams) {
	// Processes MatchProtocols.
	for _, mp := range param.MatchProtocols {
		rule.MatchProtocols = append(rule.MatchProtocols, v1.MatchProtocol(mp))

	}

	// Processes MatchPaths.
	for _, mp := range param.MatchPaths {
		rule.MatchPaths = append(rule.MatchPaths, v1.MatchPath(mp))
	}

	// Processes MatchDirectories.
	for _, md := range param.MatchDirectories {
		rule.MatchDirectories = append(rule.MatchDirectories, v1.MatchDirectory{
			Directory:  md.Directory,
			FromSource: []v1.NimbusFromSource{},
		})
	}

	// Processes MatchPatterns.
	for _, mp := range param.MatchPatterns {
		rule.MatchPatterns = append(rule.MatchPatterns, v1.MatchPattern(mp))
	}

	// Processes MatchCapabilities.
	for _, mc := range param.MatchCapabilities {
		rule.MatchCapabilities = append(rule.MatchCapabilities, v1.MatchCapability(mc))
	}

	// Processes MatchSyscalls.
	for _, ms := range param.MatchSyscalls {
		rule.MatchSyscalls = append(rule.MatchSyscalls, v1.MatchSyscall(ms))
	}

	// Processes FromCIDRSet.
	for _, fcs := range param.FromCIDRSet {
		rule.FromCIDRSet = append(rule.FromCIDRSet, v1.CIDRSet(fcs))
	}

	// Processes ToPorts.
	for _, tp := range param.ToPorts {
		var ports []v1.Port
		for _, p := range tp.Ports {
			ports = append(ports, v1.Port(p))
		}
		rule.ToPorts = append(rule.ToPorts, v1.ToPort{
			Ports: ports,
		})
	}
}

// extractSelector extracts match labels from a Selector.
func extractSelector(selector v1.Selector) (map[string]string, error) {
	matchLabels := make(map[string]string) // Initialize map for match labels.

	// Process CEL expressions.
	if len(selector.CEL) > 0 {
		celMatchLabels, err := ProcessCEL(selector.CEL)
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

// ProcessCEL processes CEL expressions to generate matchLabels.
func ProcessCEL(expressions []string) (map[string]string, error) {
	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("label", decls.NewMapType(decls.String, decls.String)), // Define label as a map of string to string.
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating CEL environment: %v", err)
	}

	matchLabels := make(map[string]string)

	for _, expr := range expressions {
		ast, issues := env.Compile(expr)
		if issues != nil && issues.Err() != nil {
			return nil, fmt.Errorf("error compiling CEL expression: %v", issues.Err())
		}

		prg, err := env.Program(ast)
		if err != nil {
			return nil, fmt.Errorf("error creating CEL program: %v", err)
		}

		out, _, err := prg.Eval(map[string]interface{}{
			"label": map[string]interface{}{},
		})
		if err != nil {
			return nil, fmt.Errorf("error evaluating CEL expression: %v", err)
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
