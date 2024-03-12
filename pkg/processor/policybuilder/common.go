// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package policybuilder

import (
	"context"
	"fmt"
	"regexp"
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

	// Parse and evaluate label expressions
	for _, expr := range expressions {
		isNegated := checkNegation(expr)
		expr = PreprocessExpression(expr)

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
				continue
			}

			if outValue, ok := out.Value().(bool); ok && outValue {
				// Mark this expression as true for at least one pod
				labels := extractLabelsFromExpression(expr, podList, isNegated)
				for k, v := range labels {
					matchLabels[k] = v
				}
			}
		}
	}
	return matchLabels, nil
}

func extractLabelsFromExpression(expr string, podList corev1.PodList, isNegated bool) map[string]string {
	labels := make(map[string]string)

	if strings.Contains(expr, "==") || strings.Contains(expr, "!=") {
		key, value := parseKeyValueExpression(expr)
		labels[key] = value
	} else if strings.Contains(expr, ".contains(") {
		key, value := parseFunctionExpression(expr, "contains")
		if key != "" && value != "" {
			labels[key] = value
		}
	} else if strings.Contains(expr, " in ") {
		key, values := parseInExpression(expr)
		for _, pod := range podList.Items {
			labelValue, exists := pod.Labels[key]
			if !exists {
				continue
			}
			if contains(values, labelValue) {
				labels[key] = labelValue
			}
		}
	} else if strings.Contains(expr, ".startsWith(") {
		labels = parseStartsWithEndsWithExpression(expr, podList, "startsWith")
	} else if strings.Contains(expr, ".endsWith(") {
		labels = parseStartsWithEndsWithExpression(expr, podList, "endsWith")
	} else if strings.Contains(expr, ".matches(") {
		labels = parseMatchesExpression(expr, podList)
	}
	if isNegated {
		labels = excludeLabels(podList, labels)
	}
	return labels
}

// Helper function to check if expression is negated (!expr) and extract clean expression.
func checkNegation(expr string) bool {
	isNegated := strings.HasPrefix(expr, "'!") || strings.HasPrefix(expr, "!") || strings.HasPrefix(expr, "!='") || strings.Contains(expr, " != ")
	return isNegated
}

func PreprocessExpression(expr string) string {
	expr = strings.TrimSpace(expr)
	expr = regexp.MustCompile(`^['"]|['"]$`).ReplaceAllString(expr, "")
	expr = strings.ReplaceAll(expr, `\"`, `"`)
	expr = strings.ReplaceAll(expr, `\'`, `'`)
	expr = strings.Replace(expr, `\"`, `"`, -1)
	expr = regexp.MustCompile(`^['"]|['"]$`).ReplaceAllString(expr, "")
	if strings.Count(expr, "\"")%2 != 0 {
		expr += "\""
	} else if strings.Count(expr, "'")%2 != 0 {
		expr += "'"
	}

	return expr
}

func parseKeyValueExpression(expr string) (string, string) {
	expr = PreprocessExpression(expr)
	var operator string
	if strings.Contains(expr, "==") {
		operator = "=="
	} else if strings.Contains(expr, "!=") {
		operator = "!="
	} else {
		return "", ""
	}

	parts := strings.SplitN(expr, operator, 2)
	if len(parts) != 2 {
		return "", ""
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	key = strings.TrimPrefix(key, "labels[")
	key = strings.TrimSuffix(key, "]")
	key = strings.Trim(key, `"'`)
	value = strings.Trim(value, `"'`)
	return key, value
}

// Parses function expressions like 'labels["key"].contains("value")'
func parseFunctionExpression(expr string, functionName string) (string, string) {
	start := strings.Index(expr, `labels["`) + len(`labels["`)
	if start == -1 {
		return "", "" // Key not found
	}
	end := strings.Index(expr[start:], `"]`)
	if end == -1 {
		return "", "" // Incorrectly formatted expression
	}
	key := expr[start : start+end]

	functionStart := strings.Index(expr, functionName+"(\"") + len(functionName+"(\"")
	functionEnd := strings.LastIndex(expr, "\")")
	if functionStart == -1 || functionEnd == -1 || functionStart >= functionEnd {
		return "", "" // Function or value not found
	}
	value := expr[functionStart:functionEnd]

	return key, value
}

func parseInExpression(expr string) (string, []string) {
	start := strings.Index(expr, `labels["`) + len(`labels["`)
	if start == -1 {
		return "", nil // Key not found
	}
	end := strings.Index(expr[start:], `"]`)
	if end == -1 {
		return "", nil // Incorrectly formatted expression
	}
	key := expr[start : start+end]

	valuesStart := strings.Index(expr, " in [") + len(" in [")
	valuesEnd := strings.LastIndex(expr, "]")
	if valuesStart == -1 || valuesEnd == -1 || valuesStart >= valuesEnd {
		return "", nil // Values not found
	}
	valuesString := expr[valuesStart:valuesEnd]
	valuesParts := strings.Split(valuesString, ",")

	var values []string
	for _, part := range valuesParts {
		value := strings.TrimSpace(part)
		value = strings.Trim(value, "\"'")
		values = append(values, value)
	}

	return key, values
}

func parseStartsWithEndsWithExpression(expr string, podList corev1.PodList, functionName string) map[string]string {
	labels := make(map[string]string)
	key, pattern := parseFunctionExpression(expr, functionName)

	for _, pod := range podList.Items {
		labelValue, exists := pod.Labels[key]
		if !exists {
			continue
		}

		var match bool
		if functionName == "startsWith" && strings.HasPrefix(labelValue, pattern) {
			match = true
		} else if functionName == "endsWith" && strings.HasSuffix(labelValue, pattern) {
			match = true
		}

		if match {
			// If a label matches, add it to the labels map
			labels[key] = labelValue
		}
	}

	return labels
}

func parseMatchesExpression(expr string, podList corev1.PodList) map[string]string {
	key, pattern := parseFunctionExpression(expr, "matches")
	labels := make(map[string]string)

	regex, _ := regexp.Compile(pattern)

	for _, pod := range podList.Items {
		labelValue, exists := pod.Labels[key]
		if !exists {
			continue
		}

		// Check if the label's value matches the pattern
		if regex.MatchString(labelValue) {
			labels[key] = labelValue
		}
	}

	return labels
}

func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
func excludeLabels(podList corev1.PodList, excludeMap map[string]string) map[string]string {
	remainingLabels := make(map[string]string)

	// Iterate through all pods in the namespace
	for _, pod := range podList.Items {
		// Check if the pod should be excluded based on the provided labels
		exclude := false
		for excludeKey, excludeValue := range excludeMap {
			podLabelValue, exists := pod.Labels[excludeKey]
			if exists && podLabelValue == excludeValue {
				exclude = true
				break
			}
		}

		// If the pod is not excluded, add its labels to the remainingLabels map
		if !exclude {
			for labelKey, labelValue := range pod.Labels {
				// Exclude pod-template-hash labels by default
				if labelKey != "pod-template-hash" {
					remainingLabels[labelKey] = labelValue
				}
			}
		}
	}

	return remainingLabels
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
