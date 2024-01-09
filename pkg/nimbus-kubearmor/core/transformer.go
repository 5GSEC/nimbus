// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package transformer

import (
	"context"
	"strings"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PolicyTransformer는 NimbusPolicy를 KubeArmorPolicy로 변환하는데 사용됩니다.
type PolicyTransformer struct {
	Client client.Client
}

// NewPolicyTransformer는 PolicyTransformer의 새 인스턴스를 생성합니다.
func NewPolicyTransformer(client client.Client) *PolicyTransformer {
	return &PolicyTransformer{Client: client}
}

// Transform is used to convert a NimbusPolicy to a KubeArmorPolicy
func (pt *PolicyTransformer) Transform(ctx context.Context, nimbusPolicy v1.NimbusPolicy) (*kubearmorv1.KubeArmorPolicy, error) {
	log := log.FromContext(ctx)
	log.Info("Start Converting a NimbusPolicy", "PolicyName", nimbusPolicy.Name)

	kubeArmorPolicy := &kubearmorv1.KubeArmorPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nimbusPolicy.Name,
			Namespace: nimbusPolicy.Namespace,
		},
		Spec: kubearmorv1.KubeArmorPolicySpec{
			// Selector 설정
			Selector: kubearmorv1.SelectorType{
				MatchLabels: nimbusPolicy.Spec.Selector.MatchLabels,
			},
		},
	}
	// NimbusPolicy의 Selector를 KubeArmorPolicy의 Selector로 변환
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
			Protocol: "raw", // 적절한 기본 프로토콜을 설정합니다.
		})
	}
	if len(kubeArmorPolicy.Spec.Capabilities.MatchCapabilities) == 0 {
		kubeArmorPolicy.Spec.Capabilities.MatchCapabilities = append(kubeArmorPolicy.Spec.Capabilities.MatchCapabilities, kubearmorv1.MatchCapabilitiesType{
			Capability: "lease", // 적절한 기본 능력을 설정합니다.
		})
	}

	return kubeArmorPolicy, nil
}

// ApplyPolicy는 주어진 KubeArmorPolicy를 적용하거나 업데이트합니다.
func (pt *PolicyTransformer) ApplyPolicy(ctx context.Context, kubeArmorPolicy *kubearmorv1.KubeArmorPolicy) error {
	log := log.FromContext(ctx)
	log.Info("Apply KubeArmorPolicy", "PolicyName", kubeArmorPolicy.Name)

	// 이미 존재하는 KubeArmorPolicy 가져오기
	existingPolicy := &kubearmorv1.KubeArmorPolicy{}
	err := pt.Client.Get(ctx, types.NamespacedName{Name: kubeArmorPolicy.Name, Namespace: kubeArmorPolicy.Namespace}, existingPolicy)
	if err != nil && !errors.IsNotFound(err) {
		log.Error(err, "Existing KubeArmorPolicy lookup failed", "PolicyName", kubeArmorPolicy.Name)
		return err
	}

	// 존재하는 경우 업데이트, 그렇지 않은 경우 생성
	if errors.IsNotFound(err) {
		log.Info("Create a KubeArmorPolicy", "PolicyName", kubeArmorPolicy.Name)
		err := pt.Client.Create(ctx, kubeArmorPolicy)
		if err != nil {
			log.Error(err, "KubeArmorPolicy Creation Failed", "PolicyName", kubeArmorPolicy.Name)
			return err
		}
	} else {
		log.Info("Update the KubeArmorPolicy", "PolicyName", kubeArmorPolicy.Name)
		existingPolicy.Spec = kubeArmorPolicy.Spec
		err := pt.Client.Update(ctx, existingPolicy)
		if err != nil {
			log.Error(err, "KubeArmorPolicy Update Failed", "PolicyName", kubeArmorPolicy.Name)
			return err
		}
	}

	return nil
}
