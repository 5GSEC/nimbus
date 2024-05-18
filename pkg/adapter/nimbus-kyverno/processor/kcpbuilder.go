// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"strings"

	v1 "github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-kyverno/utils"
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
			kcp.Name = cnp.Name + "-" + strings.ToLower(id)
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
		return clusterEscapeToHost(cnp, cnp.Spec.NimbusRules[0].Rule)
	default:
		return kyvernov1.ClusterPolicy{}
	}
}

func clusterEscapeToHost(cnp *v1.ClusterNimbusPolicy, rule v1.Rule) kyvernov1.ClusterPolicy {
	var psa_level api.Level = api.LevelBaseline

	if rule.Params["psa_level"] != nil {

		switch rule.Params["psa_level"][0] {
		case "restricted":
			psa_level = api.LevelRestricted

		case "privileged":
			psa_level = api.LevelPrivileged

		default:
			psa_level = api.LevelBaseline
		}

	}

	var matchFilters, excludeFilters []kyvernov1.ResourceFilter
	var resourceFilter kyvernov1.ResourceFilter

	if len(cnp.Spec.NsSelector.MatchNames) > 0 {
		resourceFilter = kyvernov1.ResourceFilter{
			ResourceDescription: kyvernov1.ResourceDescription{
				Kinds: []string{
					"v1/Pod",
				},
				Namespaces: cnp.Spec.NsSelector.MatchNames,
				Selector: &metav1.LabelSelector{
					MatchLabels: cnp.Spec.ObjSelector.MatchLabels,
				},
			},
		}
		matchFilters = append(matchFilters, resourceFilter)
	}

	if len(cnp.Spec.NsSelector.ExcludeNames) > 0 {
		resourceFilter = kyvernov1.ResourceFilter{
			ResourceDescription: kyvernov1.ResourceDescription{
				Namespaces: cnp.Spec.NsSelector.ExcludeNames,
			},
		}
		excludeFilters = append(excludeFilters, resourceFilter)
	}

	background := true
	return kyvernov1.ClusterPolicy{
		Spec: kyvernov1.Spec{
			Background: &background,
			Rules: []kyvernov1.Rule{
				{
					Name: "restricted",
					MatchResources: kyvernov1.MatchResources{
						Any: matchFilters,
					},
					ExcludeResources: kyvernov1.MatchResources{
						Any: excludeFilters,
					},
					Validation: kyvernov1.Validation{
						PodSecurity: &kyvernov1.PodSecurity{
							Level:   psa_level,
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
