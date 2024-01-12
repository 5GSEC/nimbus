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

func BuildClusterNimbusPolicy(ctx context.Context, client client.Client, scheme *runtime.Scheme, clusterBindingInfo *intentbinder.BindingInfo) (*v1.ClusterNimbusPolicy, error) {
	logger := log.FromContext(ctx)
	logger.Info("Building ClusterNimbusPolicy")

	var nimbusRules []v1.NimbusRules
	for _, intentName := range clusterBindingInfo.IntentNames {
		intent, err := FetchIntentByName(ctx, client, intentName)
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

	binding, err := fetchClusterBinding(ctx, client, clusterBindingInfo.BindingNames[0])
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

	if err = ctrl.SetControllerReference(&binding, clusterNimbusPolicy, scheme); err != nil {
		logger.Error(err, "failed to set OwnerReference")
		return nil, err
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

func fetchClusterBinding(ctx context.Context, client client.Client, clusterBindingName string) (v1.ClusterSecurityIntentBinding, error) {
	logger := log.FromContext(ctx)
	var clusterBinding v1.ClusterSecurityIntentBinding
	if err := client.Get(ctx, types.NamespacedName{Name: clusterBindingName}, &clusterBinding); err != nil {
		logger.Error(err, "failed to get ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding", clusterBindingName)
		return v1.ClusterSecurityIntentBinding{}, err
	}
	return clusterBinding, nil
}
