// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package policybuilder

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/processor/intentbinder"
)

// BuildNimbusPolicy generates a NimbusPolicy based on given
// SecurityIntentBinding.
func BuildNimbusPolicy(ctx context.Context, logger logr.Logger, k8sClient client.Client, scheme *runtime.Scheme, sib v1.SecurityIntentBinding) (*v1.NimbusPolicy, error) {
	logger.Info("Building NimbusPolicy")

	intents := intentbinder.ExtractIntents(ctx, k8sClient, &sib)
	if len(intents) == 0 {
		return nil, fmt.Errorf("no SecurityIntents found in the cluster")
	}

	var nimbusRules []v1.NimbusRules
	for _, intent := range intents {
		nimbusRules = append(nimbusRules, v1.NimbusRules{
			ID:          intent.Spec.Intent.ID,
			Description: intent.Spec.Intent.Description,
			Rule: v1.Rule{
				RuleAction: intent.Spec.Intent.Action,
				Params:     intent.Spec.Intent.Params,
			},
		})
	}

	matchLabels, err := extractSelector(ctx, k8sClient, sib.Namespace, sib.Spec.Selector)
	if err != nil {
		return nil, err
	}
	if len(matchLabels) == 0 {
		return nil, nil
	}

	nimbusPolicy := &v1.NimbusPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sib.Name,
			Namespace: sib.Namespace,
			Labels:    sib.Labels,
		},
		Spec: v1.NimbusPolicySpec{
			Selector: v1.NimbusSelector{
				MatchLabels: matchLabels,
			},
			NimbusRules: nimbusRules,
		},
	}

	if err = ctrl.SetControllerReference(&sib, nimbusPolicy, scheme); err != nil {
		return nil, errors.Wrap(err, "failed to set NimbusPolicy OwnerReference")
	}

	logger.Info("NimbusPolicy built successfully", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)
	return nimbusPolicy, nil
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
