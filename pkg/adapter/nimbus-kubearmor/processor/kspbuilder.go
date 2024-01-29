// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"io"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
)

func BuildKspsFrom(logger logr.Logger, np *v1.NimbusPolicy) []kubearmorv1.KubeArmorPolicy {
	// Build KSPs based on given IDs
	var ksps []kubearmorv1.KubeArmorPolicy
	for _, nimbusRule := range np.Spec.NimbusRules {
		id := nimbusRule.ID
		if idpool.IsIdSupportedBy(id, "kubearmor") {
			ksp := buildKspFor(id)
			ksp.Name = np.Name + "-" + strings.ToLower(id)
			ksp.Namespace = np.Namespace
			ksp.Spec.Message = nimbusRule.Description
			ksp.Spec.Selector.MatchLabels = np.Spec.Selector.MatchLabels
			ksp.Spec.Action = kubearmorv1.ActionType(nimbusRule.Rule.RuleAction)
			processRuleParams(&ksp, nimbusRule.Rule)
			addManagedByAnnotation(&ksp)
			ksps = append(ksps, ksp)
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
	default:
		return kubearmorv1.KubeArmorPolicy{}
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
	var ksp kubearmorv1.KubeArmorPolicy
	fileUrl := "https://raw.githubusercontent.com/kubearmor/policy-templates/main/nist/system/ksp-nist-si-4-execute-package-management-process-in-container.yaml"
	response, err := http.Get(fileUrl)

	if err != nil {
		return ksp
	}
	defer func() {
		err = response.Body.Close()
		if err != nil {
			return
		}
	}()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return ksp
	}
	_ = yaml.Unmarshal(data, &ksp)

	// remove explicit action
	ksp.Spec.Process.Action = ""
	return ksp
}

func addManagedByAnnotation(ksp *kubearmorv1.KubeArmorPolicy) {
	ksp.Annotations = make(map[string]string)
	ksp.Annotations["app.kubernetes.io/managed-by"] = "nimbus-kubearmor"
}
