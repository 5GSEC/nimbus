// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"strings"

	v1 "github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
	"github.com/go-logr/logr"
	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/pod-security-admission/api"
)

func BuildKpsFrom(logger logr.Logger, np *v1.NimbusPolicy) []kyvernov1.Policy {
	// Build KPs based on given IDs
	var kps []kyvernov1.Policy
	for _, nimbusRule := range np.Spec.NimbusRules {
		id := nimbusRule.ID
		if idpool.IsIdSupportedBy(id, "kyverno") {
			kp := buildKpFor(id, np)
			kp.Name = np.Name + "-" + strings.ToLower(id)
			kp.Namespace = np.Namespace
			kp.Annotations = make(map[string]string)
			kp.Annotations["policies.kyverno.io/description"] = nimbusRule.Description
			if nimbusRule.Rule.RuleAction == "Block" {
				kp.Spec.ValidationFailureAction = kyvernov1.ValidationFailureAction("Enforce")
			} else {
				kp.Spec.ValidationFailureAction = kyvernov1.ValidationFailureAction("Audit")
			}
			addManagedByAnnotation(&kp)
			kps = append(kps, kp)
		} else {
			logger.Info("Kyverno does not support this ID", "ID", id,
				"NimbusPolicy", np.Name, "NimbusPolicy.Namespace", np.Namespace)
		}
	}
	return kps
}

// buildKpFor builds a KyvernoPolicy based on intent ID supported by Kyverno Policy Engine.
func buildKpFor(id string, np *v1.NimbusPolicy) kyvernov1.Policy {
	switch id {
	case idpool.EscapeToHost:
		return escapeToHost(np, np.Spec.NimbusRules[0].Rule)
	default:
		return kyvernov1.Policy{}
	}
}

func escapeToHost(np *v1.NimbusPolicy, rule v1.Rule) kyvernov1.Policy {
	background := true
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

	labels := np.Spec.Selector.MatchLabels
	lis := rule.Params["exclude_resources"]
	exclusionLables := make(map[string]string)
	for _, item := range lis {
		parts := strings.Split(item, ":")
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			exclusionLables[key] = value
		}
	}

	kp := kyvernov1.Policy{
		Spec: kyvernov1.Spec{
			Background: &background,
			Rules: []kyvernov1.Rule{
				{
					Name: "restricted",
					MatchResources: kyvernov1.MatchResources{
						Any: kyvernov1.ResourceFilters{
							kyvernov1.ResourceFilter{
								ResourceDescription: kyvernov1.ResourceDescription{
									Kinds: []string{
										"v1/Pod",
									},
									Selector: &metav1.LabelSelector{
										MatchLabels: np.Spec.Selector.MatchLabels,
									},
								},
							},
						},
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

	if len(labels) > 0 {
		kp.Spec.Rules[0].MatchResources.Any[0].ResourceDescription.Selector.MatchLabels = labels
	}

	return kp
}

func addManagedByAnnotation(kp *kyvernov1.Policy) {
	kp.Annotations["app.kubernetes.io/managed-by"] = "nimbus-kyverno"
}
