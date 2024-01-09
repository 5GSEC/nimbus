// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package converter

import (
	"context"
	"strings"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PolicyConverter struct {
	Client client.Client
}

func NewPolicyConverter(client client.Client) *PolicyConverter {
	return &PolicyConverter{Client: client}
}

func (pt *PolicyConverter) Converter(ctx context.Context, nimbusPolicy v1.NimbusPolicy) (*kubearmorv1.KubeArmorPolicy, error) {
	log := log.FromContext(ctx)
	log.Info("Start Converting a NimbusPolicy", "PolicyName", nimbusPolicy.Name)

	kubeArmorPolicy := &kubearmorv1.KubeArmorPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nimbusPolicy.Name,
			Namespace: nimbusPolicy.Namespace,
		},
		Spec: kubearmorv1.KubeArmorPolicySpec{
			Selector: kubearmorv1.SelectorType{
				MatchLabels: nimbusPolicy.Spec.Selector.MatchLabels,
			},
		},
	}
	kubeArmorPolicy.Spec.Selector.MatchLabels = nimbusPolicy.Spec.Selector.MatchLabels

	for _, nimbusRule := range nimbusPolicy.Spec.NimbusRules {
		idParts := strings.Split(nimbusRule.Id, "-")
		if len(idParts) != 3 {
			log.Info("Invalid rule ID format", "ID", nimbusRule.Id)
			continue
		}

		ruleType := idParts[1]
		category := idParts[2]

		for _, rule := range nimbusRule.Rule {
			kubeArmorPolicy.Spec.Action = kubearmorv1.ActionType(rule.RuleAction)

			switch ruleType {
			case "proc":
				if processType, err := handleProcessPolicy(rule, category); err == nil {
					kubeArmorPolicy.Spec.Process = processType
				} else {
					log.Error(err, "Failed to handle process policy")
					return nil, err
				}

			case "file":
				if fileType, err := handleFilePolicy(rule, category); err == nil {
					kubeArmorPolicy.Spec.File = fileType
				} else {
					log.Error(err, "Failed to handle file policy")
					return nil, err
				}

			case "net":
				if networkType, err := handleNetworkPolicy(rule); err == nil {
					kubeArmorPolicy.Spec.Network = networkType
				} else {
					log.Error(err, "Failed to handle network policy")
					return nil, err
				}

			case "syscall":
				if syscallType, err := handleSyscallPolicy(rule, category); err == nil {
					kubeArmorPolicy.Spec.Syscalls = syscallType
				} else {
					log.Error(err, "Failed to handle syscall policy")
					return nil, err
				}

			case "cap":
				if capabilityType, err := handleCapabilityPolicy(rule); err == nil {
					kubeArmorPolicy.Spec.Capabilities = capabilityType
				} else {
					log.Error(err, "Failed to handle capability policy")
					return nil, err
				}
			default:
				log.Info("Unsupported rule type", "Type", ruleType)
			}
		}
	}

	if len(kubeArmorPolicy.Spec.Network.MatchProtocols) == 0 {
		kubeArmorPolicy.Spec.Network.MatchProtocols = append(kubeArmorPolicy.Spec.Network.MatchProtocols, kubearmorv1.MatchNetworkProtocolType{
			Protocol: "raw",
		})
	}
	if len(kubeArmorPolicy.Spec.Capabilities.MatchCapabilities) == 0 {
		kubeArmorPolicy.Spec.Capabilities.MatchCapabilities = append(kubeArmorPolicy.Spec.Capabilities.MatchCapabilities, kubearmorv1.MatchCapabilitiesType{
			Capability: "lease",
		})
	}

	return kubeArmorPolicy, nil
}
