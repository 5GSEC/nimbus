// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"encoding/json"
	"strings"

	v1alpha1 "github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
	"github.com/go-logr/logr"
	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildKcpsFrom(logger logr.Logger, cnp *v1alpha1.ClusterNimbusPolicy) []kyvernov1.ClusterPolicy {
	// Build KCPs based on given IDs
	var kcps []kyvernov1.ClusterPolicy
	for _, nimbusRule := range cnp.Spec.NimbusRules {
		id := nimbusRule.ID
		if idpool.IsIdSupportedBy(id, "coco") {
			kcp := buildKcpFor(id, cnp)
			kcp.Name = cnp.Name + "-" + strings.ToLower(id)
			kcp.Annotations = make(map[string]string)
			kcp.Annotations["policies.kyverno.io/description"] = nimbusRule.Description
			if nimbusRule.Rule.RuleAction == "Block" {
				kcp.Spec.ValidationFailureAction = kyvernov1.ValidationFailureAction("Enforce")
			} else {
				kcp.Spec.ValidationFailureAction = kyvernov1.ValidationFailureAction("Audit")
			}
			kcp.Spec.MutateExistingOnPolicyUpdate = true
			addManagedByAnnotationForClusterScopedPolicy(&kcp)
			kcps = append(kcps, kcp)
		} else {
			logger.Info("Coco does not support this ID", "ID", id,
				"NimbusPolicy", cnp.Name, "NimbusPolicy.Namespace", cnp.Namespace)
		}
	}
	return kcps
}

// buildKcpFor builds a KyvernoPolicy based on intent ID supported by Kyverno Policy Engine.
func buildKcpFor(id string, cnp *v1alpha1.ClusterNimbusPolicy) kyvernov1.ClusterPolicy {
	switch id {
	case idpool.CocoWorkload:
		return clusterCocoWorkload(cnp)
	default:
		return kyvernov1.ClusterPolicy{}
	}
}

var nsBlackList = []string{"kube-system"}

func clusterCocoWorkload(cnp *v1alpha1.ClusterNimbusPolicy) kyvernov1.ClusterPolicy {
	var matchResource kyvernov1.ResourceDescription
	var excludeFilters []kyvernov1.ResourceFilter

	// exclude kube-system
	excludeFilters = append(excludeFilters, kyvernov1.ResourceFilter{
		ResourceDescription: kyvernov1.ResourceDescription{
			Namespaces: nsBlackList,
		},
	})

	if len(cnp.Spec.NsSelector.MatchNames) > 0 {
		matchResource = kyvernov1.ResourceDescription{
			Kinds: []string{
				"Deployment",
			},
			Namespaces: cnp.Spec.NsSelector.MatchNames,
		}
		if len(cnp.Spec.WorkloadSelector.MatchLabels) > 0 {
			matchResource.Selector = &metav1.LabelSelector{
				MatchLabels: cnp.Spec.WorkloadSelector.MatchLabels,
			}
		}
	} else if len(cnp.Spec.NsSelector.ExcludeNames) > 0 {
		matchResource = kyvernov1.ResourceDescription{
			Kinds: []string{
				"Deployment",
			},
		}
		excludeFilters = append(excludeFilters, kyvernov1.ResourceFilter{
			ResourceDescription: kyvernov1.ResourceDescription{
				Namespaces: cnp.Spec.NsSelector.ExcludeNames,
			},
		})
	}

	background := true

	targets := []kyvernov1.TargetResourceSpec{
		{
			ResourceSpec: kyvernov1.ResourceSpec{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
		},
	}

	patchStrategicMerge := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"runtimeClassName": "kata-qemu-snp",
				},
			},
		},
	}
	patchBytes, _ := json.Marshal(patchStrategicMerge)

	return kyvernov1.ClusterPolicy{
		Spec: kyvernov1.Spec{
			Background: &background,
			Rules: []kyvernov1.Rule{
				{
					Name: "coco-workload",
					MatchResources: kyvernov1.MatchResources{
						ResourceDescription: matchResource,
					},
					ExcludeResources: kyvernov1.MatchResources{
						Any: excludeFilters,
					},
					Mutation: kyvernov1.Mutation{
						Targets:                targets,
						RawPatchStrategicMerge: &apiextv1.JSON{Raw: patchBytes},
					},
				},
			},
		},
	}
}

func addManagedByAnnotationForClusterScopedPolicy(kcp *kyvernov1.ClusterPolicy) {
	kcp.Annotations["app.kubernetes.io/managed-by"] = "nimbus-coco"
}
