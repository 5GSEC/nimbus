// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"strings"

	v1 "github.com/5GSEC/nimbus/api/v1alpha"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
	"github.com/go-logr/logr"
	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/pod-security-admission/api"
)

func BuildKcpsFrom(logger logr.Logger, cnp *v1.ClusterNimbusPolicy) []kyvernov1.ClusterPolicy {
	// Build KCPs based on given IDs
	var kcps []kyvernov1.ClusterPolicy
	for _, nimbusRule := range cnp.Spec.NimbusRules {
		id := nimbusRule.ID
		if idpool.IsIdSupportedBy(id, "kyverno") {
			kcp := buildKcpFor(id, cnp)
			kcp.Name = cnp.Name + "-" + strings.ToLower(id) + "-" + strings.ToLower(id)
			kcp.Annotations = make(map[string]string)
			kcp.Annotations["policies.kyverno.io/description"] = nimbusRule.Description
			if nimbusRule.Rule.RuleAction == "Block" {
				kcp.Spec.ValidationFailureAction = kyvernov1.ValidationFailureAction("Enforce")
			} else {
				kcp.Spec.ValidationFailureAction = kyvernov1.ValidationFailureAction("Audit")
			}
			addManagedByAnnotationForClusterScopedPolicy(&kcp)
			kcps = append(kcps, kcp)
		} else {
			logger.Info("Kyverno does not support this ID", "ID", id,
				"NimbusPolicy", cnp.Name, "NimbusPolicy.Namespace", cnp.Namespace)
		}
	}
	return kcps
}

// buildKpFor builds a KyvernoPolicy based on intent ID supported by Kyverno Policy Engine.
func buildKcpFor(id string, cnp *v1.ClusterNimbusPolicy) kyvernov1.ClusterPolicy {
	switch id {
	case idpool.EscapeToHost:
		return clusterEscapeToHost(cnp)
	default:
		return kyvernov1.ClusterPolicy{}
	}
}

func clusterEscapeToHost(cnp *v1.ClusterNimbusPolicy) kyvernov1.ClusterPolicy {

	labelsPerNamespace := make(map[string]map[string]string) //todo: what if we want to apply policy to  multiple resources in different namespaces?

	// Function to add or update values for a key
	addOrUpdate := func(key string, innerMap map[string]string) {
		if val, ok := labelsPerNamespace[key]; ok {
			// Key exists, update the inner map
			for k, v := range innerMap {
				val[k] = v
			}
			labelsPerNamespace[key] = val
		} else {
			// Key does not exist, add a new entry
			labelsPerNamespace[key] = innerMap
		}
	}

	for _, resource := range cnp.Spec.Selector.Resources {
		namespace := resource.Namespace
		addOrUpdate(namespace, resource.MatchLabels)
	}

	var resourceFilters []kyvernov1.ResourceFilter
	var resourceFilter kyvernov1.ResourceFilter

	for namespace, labels := range labelsPerNamespace {
		resourceFilter = kyvernov1.ResourceFilter{
			ResourceDescription: kyvernov1.ResourceDescription{
				Kinds: []string{
					"v1/Pod",
				},
				Namespaces: []string{
					namespace,
				},
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
			},
		}

		resourceFilters = append(resourceFilters, resourceFilter)
	}

	background := true
	return kyvernov1.ClusterPolicy{
		Spec: kyvernov1.Spec{
			Background: &background,
			Rules: []kyvernov1.Rule{
				{
					Name: "restricted",
					MatchResources: kyvernov1.MatchResources{
						Any: resourceFilters,
					},
					Validation: kyvernov1.Validation{
						PodSecurity: &kyvernov1.PodSecurity{
							Level:   api.LevelRestricted,
							Version: "latest",
						},
					},
				},
			},
		},
	}
}

func addManagedByAnnotationForClusterScopedPolicy(kcp *kyvernov1.ClusterPolicy) {
	kcp.Annotations["app.kubernetes.io/managed-by"] = "nimbus-kyverno"
}
