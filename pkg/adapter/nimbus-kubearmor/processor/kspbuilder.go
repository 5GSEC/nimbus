// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"strings"

	"github.com/go-logr/logr"
	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"

	v1 "github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
)

func BuildKspsFrom(logger logr.Logger, np *v1.NimbusPolicy) []kubearmorv1.KubeArmorPolicy {
	// Build KSPs based on given IDs
	var ksps []kubearmorv1.KubeArmorPolicy
	var ksp kubearmorv1.KubeArmorPolicy
	for _, nimbusRule := range np.Spec.NimbusRules {
		id := nimbusRule.ID
		if idpool.IsIdSupportedBy(id, "kubearmor") {
			if _, ok := idpool.KaIDPolicies[id]; ok {
				for _, policyName := range idpool.KaIDPolicies[id] {
					ksp = buildKspFor(policyName)
					ksp.Name = np.Name + "-" + strings.ToLower(id) + "-" + strings.ToLower(policyName)
					ksp.Namespace = np.Namespace
					ksp.Spec.Message = nimbusRule.Description
					ksp.Spec.Selector.MatchLabels = np.Spec.Selector.MatchLabels
					ksp.Spec.Action = kubearmorv1.ActionType(nimbusRule.Rule.RuleAction)
					processRuleParams(&ksp, nimbusRule.Rule)
					addManagedByAnnotation(&ksp)
					ksps = append(ksps, ksp)
				}
			} else {
				ksp = buildKspFor(id)
				ksp.Name = np.Name + "-" + strings.ToLower(id)
				ksp.Namespace = np.Namespace
				ksp.Spec.Message = nimbusRule.Description
				ksp.Spec.Selector.MatchLabels = np.Spec.Selector.MatchLabels
				ksp.Spec.Action = kubearmorv1.ActionType(nimbusRule.Rule.RuleAction)
				processRuleParams(&ksp, nimbusRule.Rule)
				addManagedByAnnotation(&ksp)
				ksps = append(ksps, ksp)
			}
		} else {
			logger.Info("KubeArmor does not support this ID", "ID", id,
				"NimbusPolicy", np.Name, "NimbusPolicy.Namespace", np.Namespace)
		}
	}
	return ksps
}

// buildKspFor builds a KubeArmorPolicy based on intent ID supported by KubeArmor Security Engine.
func buildKspFor(id string) kubearmorv1.KubeArmorPolicy {
	switch id {
	case idpool.SwDeploymentTools:
		return swDeploymentToolsKsp()
	case idpool.UnAuthorizedSaTokenAccess:
		return unAuthorizedSaTokenAccessKsp()
	case idpool.DNSManipulation:
		return dnsManipulationKsp()
	case idpool.DisallowChRoot:
		return disallowChRoot()
	case idpool.DisallowCapabilities:
		return disallowCapabilities()
	default:
		return kubearmorv1.KubeArmorPolicy{}
	}
}

func dnsManipulationKsp() kubearmorv1.KubeArmorPolicy {
	return kubearmorv1.KubeArmorPolicy{
		Spec: kubearmorv1.KubeArmorPolicySpec{
			File: kubearmorv1.FileType{
				MatchPaths: []kubearmorv1.FilePathType{
					{
						Path:     "/etc/resolv.conf",
						ReadOnly: true,
					},
				},
			},
		},
	}
}

func unAuthorizedSaTokenAccessKsp() kubearmorv1.KubeArmorPolicy {
	return kubearmorv1.KubeArmorPolicy{
		Spec: kubearmorv1.KubeArmorPolicySpec{
			File: kubearmorv1.FileType{
				MatchDirectories: []kubearmorv1.FileDirectoryType{
					{
						Directory: "/run/secrets/kubernetes.io/serviceaccount/",
						Recursive: true,
					},
				},
			},
		},
	}
}

func swDeploymentToolsKsp() kubearmorv1.KubeArmorPolicy {
	return kubearmorv1.KubeArmorPolicy{
		Spec: kubearmorv1.KubeArmorPolicySpec{
			Process: kubearmorv1.ProcessType{
				MatchPaths: []kubearmorv1.ProcessPathType{
					{
						Path: "/usr/bin/apt",
					},
					{
						Path: "/usr/bin/apt-get",
					},
					{
						Path: "/sbin/apk",
					},
					{
						Path: "/bin/apt-get",
					},
					{
						Path: "/bin/apt",
					},
					{
						Path: "/usr/bin/dpkg",
					},
					{
						Path: "/bin/dpkg",
					},
					{
						Path: "/usr/bin/gdebi",
					},
					{
						Path: "/bin/gdebi",
					},
					{
						Path: "/usr/bin/make",
					},
					{
						Path: "/bin/make",
					},
					{
						Path: "/usr/bin/yum",
					},
					{
						Path: "/bin/yum",
					},
					{
						Path: "/usr/bin/rpm",
					},
					{
						Path: "/bin/rpm",
					},
					{
						Path: "/usr/bin/dnf",
					},
					{
						Path: "/bin/dnf",
					},
					{
						Path: "/usr/bin/pacman",
					},
					{
						Path: "/usr/sbin/pacman",
					},
					{
						Path: "/bin/pacman",
					},
					{
						Path: "/sbin/pacman",
					},
					{
						Path: "/usr/bin/makepkg",
					},
					{
						Path: "/usr/sbin/makepkg",
					},
					{
						Path: "/bin/makepkg",
					},
					{
						Path: "/sbin/makepkg",
					},
					{
						Path: "/usr/bin/yaourt",
					},
					{
						Path: "/usr/sbin/yaourt",
					},
					{
						Path: "/bin/yaourt",
					},
					{
						Path: "/sbin/yaourt",
					},
					{
						Path: "/usr/bin/zypper",
					},
					{
						Path: "/bin/zypper",
					},
					{
						Path: "/usr/bin/curl",
					},
					{
						Path: "/bin/curl",
					},
					{
						Path: "/usr/local/bin/curl",
					},
					{
						Path: "/usr/bin/wget",
					},
					{
						Path: "/bin/wget",
					},
					{
						Path: "/usr/local/bin/curl",
					},
				},
			},
		},
	}
}

func disallowCapabilities() kubearmorv1.KubeArmorPolicy {
	return kubearmorv1.KubeArmorPolicy{
		Spec: kubearmorv1.KubeArmorPolicySpec{
			Capabilities: kubearmorv1.CapabilitiesType{
				MatchCapabilities: []kubearmorv1.MatchCapabilitiesType{
					{
						Capability: "sys_admin",
					},
					{
						Capability: "sys_ptrace",
					},
					{
						Capability: "sys_module",
					},
					{
						Capability: "dac_read_search",
					},
					{
						Capability: "dac_override",
					},
				},
			},
		},
	}
}

func disallowChRoot() kubearmorv1.KubeArmorPolicy {
	return kubearmorv1.KubeArmorPolicy{
		Spec: kubearmorv1.KubeArmorPolicySpec{
			Process: kubearmorv1.ProcessType{
				MatchPaths: []kubearmorv1.ProcessPathType{
					{
						Path: "/usr/sbin/chroot",
					},
					{
						Path: "/sbin/chroot",
					},
				},
			},
		},
	}
}

func addManagedByAnnotation(ksp *kubearmorv1.KubeArmorPolicy) {
	ksp.Annotations = make(map[string]string)
	ksp.Annotations["app.kubernetes.io/managed-by"] = "nimbus-kubearmor"
}
