// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"strings"

	v1 "github.com/5GSEC/nimbus/api/v1"
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
			kcp := buildKcpFor(id, cnp, nimbusRule.Rule)
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
func buildKcpFor(id string, cnp *v1.ClusterNimbusPolicy, rule v1.Rule) kyvernov1.ClusterPolicy {
	switch id {
	case idpool.EscapeToHost:
		return clusterEscapeToHost(cnp, rule)
	default:
		return kyvernov1.ClusterPolicy{}
	}
}

func clusterEscapeToHost(cnp *v1.ClusterNimbusPolicy, rule v1.Rule) kyvernov1.ClusterPolicy {
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

	var resourceFilters []kyvernov1.ResourceFilter
	var exclusionFilters []kyvernov1.ResourceFilter
	var kinds []string

	for _,resource := range cnp.Spec.Selector.Resources {
		kind := resource.Kind
		name := resource.Name
		switch kind {
		case "Namespace":			
			kinds = append(kinds, utils.GetGVK("deployment"), utils.GetGVK("replicaset"), utils.GetGVK("statefulset"), utils.GetGVK("pod"), utils.GetGVK("daemonset"), utils.GetGVK("configmap"), utils.GetGVK("secret"), utils.GetGVK("serviceaccount"))
			resourceFilterForNamespace := kyvernov1.ResourceFilter{
				ResourceDescription: kyvernov1.ResourceDescription{
					Kinds: kinds,
					Namespaces: []string{
						name,
					},
				},
			}
			
			if len(lis)>0 {
				excludeFilterForNamespace := kyvernov1.ResourceFilter{
					ResourceDescription: kyvernov1.ResourceDescription{
						Kinds: kinds,
						Namespaces: []string{
							name,
						},
						Selector: &metav1.LabelSelector{
							MatchLabels: exclusionLables,
						},
					},
				}

				exclusionFilters = append(exclusionFilters, excludeFilterForNamespace)
			}
			resourceFilters = append(resourceFilters, resourceFilterForNamespace)

		default:
			gvk := utils.GetGVK(kind)
			namespace := resource.Namespace
			labels := resource.MatchLabels
			var resourceFilter kyvernov1.ResourceFilter
			if len(labels) != 0 {
				resourceFilter = kyvernov1.ResourceFilter{
					ResourceDescription: kyvernov1.ResourceDescription{
						Kinds: []string{
							gvk,
						},
						Namespaces: []string{
							namespace,
						},
						Selector: &metav1.LabelSelector{
							MatchLabels: labels,
						},
					},
				}
			} else {
				resourceFilter = kyvernov1.ResourceFilter{
					ResourceDescription: kyvernov1.ResourceDescription{
						Kinds: []string{
							gvk,
						},
						Namespaces: []string{
							namespace,
						},
					},
				}
			}

			if len(lis) > 0 && len(labels) == 0 {
				excludeFilterForNamespace := kyvernov1.ResourceFilter{
					ResourceDescription: kyvernov1.ResourceDescription{
						Kinds: []string{
							gvk,
						},
						Namespaces: []string{
							namespace,
						},
						Selector: &metav1.LabelSelector{
							MatchLabels: exclusionLables,
						},
					},
				}

				exclusionFilters = append(exclusionFilters, excludeFilterForNamespace)
			}

			resourceFilters = append(resourceFilters, resourceFilter)
			
		}
	} 

	background := true
	kcp := kyvernov1.ClusterPolicy{
		Spec: kyvernov1.Spec{
			Background: &background ,
			Rules: []kyvernov1.Rule{
				{
					Name: "restricted",
					MatchResources: kyvernov1.MatchResources{
						Any: resourceFilters,
					},
					Validation: kyvernov1.Validation{
						PodSecurity : &kyvernov1.PodSecurity{
							Level: api.LevelRestricted,
							Version: "latest",
						},
					},
				},
			},
		},
	}

	if len(exclusionFilters) > 0 {
		kcp.Spec.Rules[0].ExcludeResources  =  kyvernov1.MatchResources{
			Any: exclusionFilters,			
		}
	}

	return kcp
}

func addManagedByAnnotationForClusterScopedPolicy(kcp *kyvernov1.ClusterPolicy) {
	kcp.Annotations["app.kubernetes.io/managed-by"] = "nimbus-kyverno"
}

