// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package policybuilder

import (
	"context"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/processor/intentbinder"
)

// BuildClusterNimbusPolicy generates a ClusterNimbusPolicy based on given
// SecurityIntents and ClusterSecurityIntentBinding.
func BuildClusterNimbusPolicy(ctx context.Context, logger logr.Logger, k8sClient client.Client, scheme *runtime.Scheme, csib v1.ClusterSecurityIntentBinding) *v1.ClusterNimbusPolicy {
	logger.Info("Building ClusterNimbusPolicy")
	intents := intentbinder.ExtractIntents(ctx, k8sClient, &csib)
	if len(intents) == 0 {
		logger.Info("No SecurityIntents found in the cluster")
		return nil
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

	clusterBindingSelector := extractClusterBindingSelector(csib.Spec.Selector)
	clusterNp := &v1.ClusterNimbusPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:   csib.Name,
			Labels: csib.Labels,
		},
		Spec: v1.ClusterNimbusPolicySpec{
			Selector:    clusterBindingSelector,
			NimbusRules: nimbusRules,
		},
	}

	if err := ctrl.SetControllerReference(&csib, clusterNp, scheme); err != nil {
		logger.Error(err, "failed to set OwnerReference")
		return nil
	}

	logger.Info("ClusterNimbusPolicy built successfully", "ClusterNimbusPolicy.Name", clusterNp.Name)
	return clusterNp
}

func extractClusterBindingSelector(cwSelector v1.CwSelector) v1.CwSelector {
	// Todo: Handle CEL expression
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
