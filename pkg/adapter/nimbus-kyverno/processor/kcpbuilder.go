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
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/pod-security-admission/api"
)

func BuildKcpsFrom(logger logr.Logger, cnp *v1alpha1.ClusterNimbusPolicy) []kyvernov1.ClusterPolicy {
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
func buildKcpFor(id string, cnp *v1alpha1.ClusterNimbusPolicy) kyvernov1.ClusterPolicy {
	switch id {
	case idpool.EscapeToHost:
		return clusterEscapeToHost(cnp, cnp.Spec.NimbusRules[0].Rule)
	case idpool.CocoWorkload:
		return clusterCocoRuntimeAddition(cnp, cnp.Spec.NimbusRules[0].Rule)
	default:
		return kyvernov1.ClusterPolicy{}
	}
}

var nsBlackList = []string{"kube-system"}

func clusterCocoRuntimeAddition(cnp *v1alpha1.ClusterNimbusPolicy, rule v1alpha1.Rule) kyvernov1.ClusterPolicy {
	var matchFilters, excludeFilters []kyvernov1.ResourceFilter
	labels := cnp.Spec.WorkloadSelector.MatchLabels
	excludeNamespaces := cnp.Spec.NsSelector.ExcludeNames
	namespaces := cnp.Spec.NsSelector.MatchNames
	// exclude kube-system
	resourceFilter := kyvernov1.ResourceFilter{
		ResourceDescription: kyvernov1.ResourceDescription{
			Namespaces: nsBlackList,
		},
	}
	excludeFilters = append(excludeFilters, resourceFilter)

	if namespaces[0] != "*" && len(labels) > 0 {
		for key, value := range labels {
			resourceFilter = kyvernov1.ResourceFilter{
				ResourceDescription: kyvernov1.ResourceDescription{
					Kinds: []string{
						"apps/v1/Deployment",
					},
					Namespaces: namespaces,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							key: value,
						},
					},
				},
			}
			matchFilters = append(matchFilters, resourceFilter)
		}
	} else if namespaces[0] != "*" && len(labels) == 0 {
		resourceFilter = kyvernov1.ResourceFilter{
			ResourceDescription: kyvernov1.ResourceDescription{
				Kinds: []string{
					"apps/v1/Deployment",
				},
				Namespaces: cnp.Spec.NsSelector.MatchNames,
			},
		}
		matchFilters = append(matchFilters, resourceFilter)
	} else if namespaces[0] == "*" && len(labels) > 0 {

		if len(excludeNamespaces) > 0 {
			resourceFilter = kyvernov1.ResourceFilter{
				ResourceDescription: kyvernov1.ResourceDescription{
					Namespaces: excludeNamespaces,
				},
			}
			excludeFilters = append(excludeFilters, resourceFilter)
		}

		for key, value := range labels {
			resourceFilter = kyvernov1.ResourceFilter{
				ResourceDescription: kyvernov1.ResourceDescription{
					Kinds: []string{
						"apps/v1/Deployment",
					},
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							key: value,
						},
					},
				},
			}
			matchFilters = append(matchFilters, resourceFilter)
		}
	} else if namespaces[0] == "*" && len(labels) == 0 {
		if len(excludeNamespaces) > 0 {
			resourceFilter = kyvernov1.ResourceFilter{
				ResourceDescription: kyvernov1.ResourceDescription{
					Namespaces: excludeNamespaces,
				},
			}
			excludeFilters = append(excludeFilters, resourceFilter)
		}
		resourceFilter = kyvernov1.ResourceFilter{
			ResourceDescription: kyvernov1.ResourceDescription{
				Kinds: []string{
					"apps/v1/Deployment",
				},
			},
		}
		matchFilters = append(matchFilters, resourceFilter)
	}

	patchStrategicMerge := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"runtimeClassName": "kata-clh",
				},
			},
		},
	}
	patchBytes, err := json.Marshal(patchStrategicMerge)
	if err != nil {
		panic(err)
	}
	return kyvernov1.ClusterPolicy{
		Spec: kyvernov1.Spec{
			MutateExistingOnPolicyUpdate: true,
			Rules: []kyvernov1.Rule{
				{
					Name: "add-runtime-class-to-pods",
					MatchResources: kyvernov1.MatchResources{
						Any: matchFilters,
					},
					ExcludeResources: kyvernov1.MatchResources{
						Any: excludeFilters,
					},
					Mutation: kyvernov1.Mutation{
						Targets: []kyvernov1.TargetResourceSpec{
							{
								ResourceSpec: kyvernov1.ResourceSpec{
									APIVersion: "apps/v1",
									Kind:       "Deployment",
								},
							},
						},
						RawPatchStrategicMerge: &v1.JSON{
							Raw: patchBytes,
						},
					},
				},
			},
		},
	}
}

func clusterEscapeToHost(cnp *v1alpha1.ClusterNimbusPolicy, rule v1alpha1.Rule) kyvernov1.ClusterPolicy {
	var psa_level api.Level = api.LevelBaseline

	if rule.Params["psa_level"] != nil {

		switch rule.Params["psa_level"][0] {
		case "restricted":
			psa_level = api.LevelRestricted

		default:
			psa_level = api.LevelBaseline
		}

	}

	var matchFilters, excludeFilters []kyvernov1.ResourceFilter
	labels := cnp.Spec.WorkloadSelector.MatchLabels
	excludeNamespaces := cnp.Spec.NsSelector.ExcludeNames
	namespaces := cnp.Spec.NsSelector.MatchNames
	// exclude kube-system
	resourceFilter := kyvernov1.ResourceFilter{
		ResourceDescription: kyvernov1.ResourceDescription{
			Namespaces: nsBlackList,
		},
	}
	excludeFilters = append(excludeFilters, resourceFilter)

	if namespaces[0] != "*" && len(labels) > 0 {
		for key, value := range labels {
			resourceFilter = kyvernov1.ResourceFilter{
				ResourceDescription: kyvernov1.ResourceDescription{
					Kinds: []string{
						"v1/Pod",
					},
					Namespaces: namespaces,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							key: value,
						},
					},
				},
			}
			matchFilters = append(matchFilters, resourceFilter)
		}
	} else if namespaces[0] != "*" && len(labels) == 0 {
		resourceFilter = kyvernov1.ResourceFilter{
			ResourceDescription: kyvernov1.ResourceDescription{
				Kinds: []string{
					"v1/Pod",
				},
				Namespaces: cnp.Spec.NsSelector.MatchNames,
			},
		}
		matchFilters = append(matchFilters, resourceFilter)
	} else if namespaces[0] == "*" && len(labels) > 0 {
		if len(excludeNamespaces) > 0 {
			resourceFilter = kyvernov1.ResourceFilter{
				ResourceDescription: kyvernov1.ResourceDescription{
					Namespaces: excludeNamespaces,
				},
			}
			excludeFilters = append(excludeFilters, resourceFilter)
		}
		for key, value := range labels {
			resourceFilter = kyvernov1.ResourceFilter{
				ResourceDescription: kyvernov1.ResourceDescription{
					Kinds: []string{
						"v1/Pod",
					},
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							key: value,
						},
					},
				},
			}
			matchFilters = append(matchFilters, resourceFilter)
		}
	} else if namespaces[0] == "*" && len(labels) == 0 {

		if len(excludeNamespaces) > 0 {
			resourceFilter = kyvernov1.ResourceFilter{
				ResourceDescription: kyvernov1.ResourceDescription{
					Namespaces: excludeNamespaces,
				},
			}
			excludeFilters = append(excludeFilters, resourceFilter)
		}
		resourceFilter = kyvernov1.ResourceFilter{
			ResourceDescription: kyvernov1.ResourceDescription{
				Kinds: []string{
					"v1/Pod",
				},
			},
		}
		matchFilters = append(matchFilters, resourceFilter)
	}
	background := true
	return kyvernov1.ClusterPolicy{
		Spec: kyvernov1.Spec{
			Background: &background,
			Rules: []kyvernov1.Rule{
				{
					Name: "pod-security-standard",
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
