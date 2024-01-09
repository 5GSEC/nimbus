// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"io"
	"net/http"

	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
)

// ProcessRuleParams processes the given nimbus policy rules, generating corresponding KubeArmorPolicy rules.
func ProcessRuleParams(ksp *kubearmorv1.KubeArmorPolicy, rule v1.Rule) {
	// Process
	// Why only process rules? Nimbus Policy has only process rules that's why we're processing only those rules.
	for _, matchPath := range rule.MatchPaths {
		ksp.Spec.Process.MatchPaths = append(ksp.Spec.Process.MatchPaths, kubearmorv1.ProcessPathType{
			Path: kubearmorv1.MatchPathType(matchPath.Path),
		})
	}

	for idx, matchDir := range rule.MatchDirectories {
		ksp.Spec.Process.MatchDirectories = append(ksp.Spec.Process.MatchDirectories, kubearmorv1.ProcessDirectoryType{
			Directory: kubearmorv1.MatchDirectoryType(matchDir.Directory),
		})
		var fromSources []kubearmorv1.MatchSourceType
		for _, fromSource := range matchDir.FromSource {
			fromSources = append(fromSources, kubearmorv1.MatchSourceType{
				Path: kubearmorv1.MatchPathType(fromSource.Path),
			})
		}
		ksp.Spec.Process.MatchDirectories[idx].FromSource = fromSources
	}

	for _, matchPattern := range rule.MatchPatterns {
		ksp.Spec.Process.MatchPatterns = append(ksp.Spec.Process.MatchPatterns, kubearmorv1.ProcessPatternType{
			Pattern: matchPattern.Pattern,
		})
	}

	// Network
	for _, matchProtocol := range rule.MatchProtocols {
		ksp.Spec.Network.MatchProtocols = append(ksp.Spec.Network.MatchProtocols, kubearmorv1.MatchNetworkProtocolType{
			Protocol: kubearmorv1.MatchNetworkProtocolStringType(matchProtocol.Protocol),
		})
	}
	// Ignoring SysCalls and Capabilities
}

// BuildKspFor builds a KubeArmorPolicy based on intent ID supported by KubeArmor Security Engine.
func BuildKspFor(id string) kubearmorv1.KubeArmorPolicy {
	switch id {
	case idpool.SwDeploymentTools:
		return buildSwDeploymentToolsKsp()
	default:
		return kubearmorv1.KubeArmorPolicy{}
	}
}

func buildSwDeploymentToolsKsp() kubearmorv1.KubeArmorPolicy {
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
