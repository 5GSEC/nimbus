// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package policybuilder

import (
	"context"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
)

// FetchIntentByName fetches a SecurityIntent by its name.
func FetchIntentByName(ctx context.Context, client client.Client, name string) (*v1.SecurityIntent, error) {
	logger := log.FromContext(ctx)

	var intent v1.SecurityIntent
	if err := client.Get(ctx, types.NamespacedName{Name: name}, &intent); err != nil {
		logger.Error(err, "Failed to get SecurityIntent")
		return nil, err
	}
	return &intent, nil
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
