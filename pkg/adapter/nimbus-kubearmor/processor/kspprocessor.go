// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"

	v1 "github.com/5GSEC/nimbus/api/v1"
)

// processRuleParams processes the given nimbus policy rules, generating corresponding KubeArmorPolicy rules.
func processRuleParams(ksp *kubearmorv1.KubeArmorPolicy, rule v1.Rule) {
	// Process
	// Why only process rules? Nimbus Policy has only process rules that's why we're processing only those rules.
	//for _, matchPath := range rule.MatchPaths {
	//	ksp.Spec.Process.MatchPaths = append(ksp.Spec.Process.MatchPaths, kubearmorv1.ProcessPathType{
	//		Path: kubearmorv1.MatchPathType(matchPath.Path),
	//	})
	//}
	//
	//for idx, matchDir := range rule.MatchDirectories {
	//	ksp.Spec.Process.MatchDirectories = append(ksp.Spec.Process.MatchDirectories, kubearmorv1.ProcessDirectoryType{
	//		Directory: kubearmorv1.MatchDirectoryType(matchDir.Directory),
	//	})
	//	var fromSources []kubearmorv1.MatchSourceType
	//	for _, fromSource := range matchDir.FromSource {
	//		fromSources = append(fromSources, kubearmorv1.MatchSourceType{
	//			Path: kubearmorv1.MatchPathType(fromSource.Path),
	//		})
	//	}
	//	ksp.Spec.Process.MatchDirectories[idx].FromSource = fromSources
	//}
	//
	//for _, matchPattern := range rule.MatchPatterns {
	//	ksp.Spec.Process.MatchPatterns = append(ksp.Spec.Process.MatchPatterns, kubearmorv1.ProcessPatternType{
	//		Pattern: matchPattern.Pattern,
	//	})
	//}
	//
	//// Network
	//for _, matchProtocol := range rule.MatchProtocols {
	//	ksp.Spec.Network.MatchProtocols = append(ksp.Spec.Network.MatchProtocols, kubearmorv1.MatchNetworkProtocolType{
	//		Protocol: kubearmorv1.MatchNetworkProtocolStringType(matchProtocol.Protocol),
	//	})
	//}
	// Ignoring SysCalls and Capabilities
}
