// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package policybuilder

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
)

// ProcessCEL processes CEL expressions to generate matchLabels.
func ProcessCEL(ctx context.Context, k8sClient client.Client, namespace string, expressions []string) (map[string]string, error) {
	logger := log.FromContext(ctx)
	logger.Info("Processing CEL expressions", "Namespace", namespace)

	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("labels", decls.NewMapType(decls.String, decls.String)), // Correctly declare 'labels' variable
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating CEL environment: %v", err)
	}

	matchLabels := make(map[string]string)

	// Retrieve pod list
	var podList corev1.PodList
	if err := k8sClient.List(ctx, &podList, client.InNamespace(namespace)); err != nil {
		logger.Error(err, "Error listing pods in namespace", "Namespace", namespace)
		return nil, fmt.Errorf("error listing pods: %v", err)
	}

	// Initialize an empty map to store label expressions
	labelExpressions := make(map[string]bool)

	// Parse and evaluate label expressions
	for _, expr := range expressions {
		ast, issues := env.Compile(expr)
		if issues != nil && issues.Err() != nil {
			return nil, fmt.Errorf("error compiling CEL expression: %v", issues.Err())
		}

		prg, err := env.Program(ast)
		if err != nil {
			return nil, fmt.Errorf("error creating CEL program: %v", err)
		}

		// Evaluate CEL expression for each pod
		for _, pod := range podList.Items {
			resource := map[string]interface{}{
				"labels": pod.Labels,
			}

			out, _, err := prg.Eval(map[string]interface{}{
				"labels": resource["labels"],
			})
			if err != nil {
				logger.Info("Error evaluating CEL expression for pod", "PodName", pod.Name, "error", err.Error())
				// Instead of returning an error immediately, we log the error and continue.
				break
			}

			if outValue, ok := out.Value().(bool); ok && outValue {
				// Mark this expression as true for at least one pod
				labelExpressions[expr] = true
			}
		}
	}

	// Extract labels based on true label expressions
	for expr, isTrue := range labelExpressions {
		if isTrue {
			// Extract labels from the expression and add them to matchLabels
			labels := extractLabelsFromExpression(expr)
			for k, v := range labels {
				matchLabels[k] = v
			}
		}
	}

	return matchLabels, nil
}

// Extracts labels from a CEL expression
func extractLabelsFromExpression(expr string) map[string]string {
	// This function is simplified and can be expanded based on specific needs.
	labels := make(map[string]string)

	// Simplified extraction logic for basic "key == value" expressions
	if strings.Contains(expr, "==") {
		parts := strings.Split(expr, "==")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Handle labels["key"] pattern
			key = strings.TrimPrefix(key, `labels["`)
			key = strings.TrimSuffix(key, `"]`)

			// Remove quotes from value if present
			value = strings.Trim(value, "\"'")

			// Add the extracted label to the map
			labels[key] = value
		}
	}

	return labels
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
