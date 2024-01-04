// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package utils

/*
import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	general "github.com/5GSEC/nimbus/pkg/controllers/general"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/cilium/cilium/pkg/policy/api"
	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
)

// ---------------------------------------------------
// -------- Creation of Policy Specifications --------
// ---------------------------------------------------

// BuildKubeArmorPolicySpec creates a policy specification (either KubeArmorPolicy or KubeArmorHostPolicy)
// based on the provided SecurityIntent and the type of policy.
// BuildKubeArmorPolicySpec creates a KubeArmor policy specification based on the provided SecurityIntentBinding.
func BuildKubeArmorPolicySpec(ctx context.Context, bindingInfo *general.BindingInfo) *kubearmorv1.KubeArmorPolicy {
	log := log.FromContext(ctx)
	log.Info("Creating KubeArmorPolicy", "BindingName", bindingInfo.Binding.Name)

	intent := bindingInfo.Intent[0]

	return &kubearmorv1.KubeArmorPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingInfo.Binding.Name,
			Namespace: bindingInfo.Binding.Namespace,
		},
		Spec: kubearmorv1.KubeArmorPolicySpec{
			Selector: kubearmorv1.SelectorType{
				MatchLabels: extractMatchLabels(bindingInfo),
			},
			Process:      extractToKubeArmorPolicyProcessType(bindingInfo),
			File:         extractToKubeArmorPolicyFileType(bindingInfo),
			Capabilities: extractToKubeArmorPolicyCapabilitiesType(bindingInfo),
			Network:      extractToKubeArmorPolicyNetworkType(bindingInfo),
			// TODO: To discuss
			//Network:      convertToKubeArmorHostPolicyNetworkType(extractNetworkPolicy(intent)),
			Action: kubearmorv1.ActionType(intent.Spec.Intent.Action),
		},
	}
}

// BuildCiliumNetworkPolicySpec creates a Cilium network policy specification based on the provided BindingInfo.
func BuildCiliumNetworkPolicySpec(ctx context.Context, bindingInfo *general.BindingInfo) interface{} {
	log := log.FromContext(ctx)
	log.Info("Creating CiliumNetworkPolicy", "Name", bindingInfo.Binding.Name)

	endpointSelector := getEndpointSelector(ctx, bindingInfo)
	ingressDenyRules := getIngressDenyRules(bindingInfo)

	policy := &ciliumv2.CiliumNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingInfo.Binding.Name,
			Namespace: bindingInfo.Binding.Namespace,
		},
		Spec: &api.Rule{
			EndpointSelector: endpointSelector,
			IngressDeny:      ingressDenyRules,
		},
	}
	return policy
}

// TODO: To discuss
//func convertToKubeArmorHostPolicyNetworkType(slice []interface{}) kubearmorv1.MatchHostNetworkProtocolType {
//	var result kubearmorv1.MatchHostNetworkProtocolType
//	for _, item := range slice {
//		str, ok := item.(string)
//		if !ok {
//			continue // or appropriate error handling
//		}
//		// Requires explicit type conversion to MatchNetworkProtocolStringType
//		protocol := kubearmorv1.MatchNetworkProtocolStringType(str)
//		result.MatchProtocols = append(result.MatchProtocols, kubearmorv1.MatchNetworkProtocolType{
//			Protocol: protocol,
//		})
//	}
//	return result
//}

// --------------------------------------
// -------- Utility Functions  ----------
// --------------------------------------

// extractMatchLabelsFromBinding extracts matchLabels from the SecurityIntentBinding.
func extractMatchLabels(bindingInfo *general.BindingInfo) map[string]string {
	matchLabels := make(map[string]string)
	for _, filter := range bindingInfo.Binding.Spec.Selector.Any {
		for key, val := range filter.Resources.MatchLabels {
			matchLabels[key] = val
		}
	}

	for _, filter := range bindingInfo.Binding.Spec.Selector.All {
		for key, val := range filter.Resources.MatchLabels {
			matchLabels[key] = val
		}
	}

	// Remove 'any:', 'all:', 'cel:' prefixes from the keys
	processedLabels := make(map[string]string)
	for key, val := range matchLabels {
		processedKey := removeReservedPrefixes(key)
		processedLabels[processedKey] = val
	}

	return processedLabels
}

func extractToKubeArmorPolicyProcessType(bindingInfo *general.BindingInfo) kubearmorv1.ProcessType {
	intent := bindingInfo.Intent[0]

	var processType kubearmorv1.ProcessType
	for _, resource := range intent.Spec.Intent.Resource {
		for _, process := range resource.Process {
			for _, match := range process.MatchPaths {
				if path := match.Path; path != "" && strings.HasPrefix(path, "/") {
					processType.MatchPaths = append(processType.MatchPaths, kubearmorv1.ProcessPathType{
						Path: kubearmorv1.MatchPathType(path),
					})
				}
			}
			for _, dir := range process.MatchDirectories {
				var fromSources []kubearmorv1.MatchSourceType
				for _, source := range dir.FromSource {
					fromSources = append(fromSources, kubearmorv1.MatchSourceType{
						Path: kubearmorv1.MatchPathType(source.Path),
					})
				}
				if dir.Directory != "" || len(fromSources) > 0 {
					processType.MatchDirectories = append(processType.MatchDirectories, kubearmorv1.ProcessDirectoryType{ // Adjusted type here
						Directory:  kubearmorv1.MatchDirectoryType(dir.Directory),
						Recursive:  dir.Recursive,
						FromSource: fromSources,
					})
				}
			}
			for _, pattern := range process.MatchPatterns {
				if pattern.Pattern != "" {
					processType.MatchPatterns = append(processType.MatchPatterns, kubearmorv1.ProcessPatternType{
						Pattern: pattern.Pattern,
					})
				}
			}
		}
	}
	return processType
}

func extractToKubeArmorPolicyFileType(bindingInfo *general.BindingInfo) kubearmorv1.FileType {
	intent := bindingInfo.Intent[0]

	var fileType kubearmorv1.FileType

	for _, resource := range intent.Spec.Intent.Resource {
		for _, file := range resource.File {
			for _, path := range file.MatchPaths {
				if path.Path != "" {
					fileType.MatchPaths = append(fileType.MatchPaths, kubearmorv1.FilePathType{
						Path: kubearmorv1.MatchPathType(path.Path),
					})
				}
			}

			for _, dir := range file.MatchDirectories {
				var fromSources []kubearmorv1.MatchSourceType
				for _, source := range dir.FromSource {
					fromSources = append(fromSources, kubearmorv1.MatchSourceType{
						Path: kubearmorv1.MatchPathType(source.Path),
					})
				}
				if dir.Directory != "" || len(fromSources) > 0 {
					fileType.MatchDirectories = append(fileType.MatchDirectories, kubearmorv1.FileDirectoryType{
						Directory:  kubearmorv1.MatchDirectoryType(dir.Directory),
						Recursive:  dir.Recursive,
						FromSource: fromSources,
					})
				}
			}
		}
	}

	return fileType
}

func extractToKubeArmorPolicyCapabilitiesType(bindingInfo *general.BindingInfo) kubearmorv1.CapabilitiesType {
	var capabilitiesType kubearmorv1.CapabilitiesType
	intent := bindingInfo.Intent[0]

	if len(intent.Spec.Intent.Resource) > 0 && len(intent.Spec.Intent.Resource[0].Capabilities) > 0 {
		for _, capability := range intent.Spec.Intent.Resource[0].Capabilities {
			for _, matchCapability := range capability.MatchCapabilities {
				if matchCapability.Capability != "" {
					capabilitiesType.MatchCapabilities = append(capabilitiesType.MatchCapabilities, kubearmorv1.MatchCapabilitiesType{
						Capability: kubearmorv1.MatchCapabilitiesStringType(matchCapability.Capability),
					})
				}
			}
		}
	} else {
		capabilitiesType.MatchCapabilities = append(capabilitiesType.MatchCapabilities, kubearmorv1.MatchCapabilitiesType{
			Capability: "lease",
		})
	}
	return capabilitiesType
}

func extractToKubeArmorPolicyNetworkType(bindingInfo *general.BindingInfo) kubearmorv1.NetworkType {
	var networkType kubearmorv1.NetworkType
	intent := bindingInfo.Intent[0]

	if len(intent.Spec.Intent.Resource) > 0 && len(intent.Spec.Intent.Resource[0].Network) > 0 {
		for _, network := range intent.Spec.Intent.Resource[0].Network {
			for _, matchProtocol := range network.MatchProtocols {
				var fromSources []kubearmorv1.MatchSourceType
				for _, source := range matchProtocol.FromSource {
					fromSources = append(fromSources, kubearmorv1.MatchSourceType{
						Path: kubearmorv1.MatchPathType(source.Path),
					})
				}
				if matchProtocol.Protocol != "" {
					networkType.MatchProtocols = append(networkType.MatchProtocols, kubearmorv1.MatchNetworkProtocolType{
						Protocol:   kubearmorv1.MatchNetworkProtocolStringType(matchProtocol.Protocol),
						FromSource: fromSources,
					})
				}
			}
		}
	} else {
		networkType.MatchProtocols = append(networkType.MatchProtocols, kubearmorv1.MatchNetworkProtocolType{
			Protocol: "raw",
		})
	}
	return networkType
}

// getEndpointSelector creates an endpoint selector from the SecurityIntent.
func getEndpointSelector(ctx context.Context, bindingInfo *general.BindingInfo) api.EndpointSelector {

	matchLabels := make(map[string]string)
	/// Extract matched labels from BindingInfo
	for _, filter := range bindingInfo.Binding.Spec.Selector.Any {
		for key, val := range filter.Resources.MatchLabels {
			matchLabels[key] = val
		}
	}

	for _, filter := range bindingInfo.Binding.Spec.Selector.All {
		for key, val := range filter.Resources.MatchLabels {
			matchLabels[key] = val
		}
	}

	processedLabels := make(map[string]string)
	for key, val := range matchLabels {
		processedKey := removeReservedPrefixes(key)
		processedLabels[processedKey] = val
	}

	// Create an Endpoint Selector based on processed labels
	return api.NewESFromMatchRequirements(processedLabels, nil)
}

func removeReservedPrefixes(key string) string {
	for _, prefix := range []string{"any:", "all:", "cel:"} {
		for strings.HasPrefix(key, prefix) {
			key = strings.TrimPrefix(key, prefix)
		}
	}
	return strings.TrimSpace(key)
}

// getIngressDenyRules generates ingress deny rules from SecurityIntent specified in BindingInfo.
func getIngressDenyRules(bindingInfo *general.BindingInfo) []api.IngressDenyRule {
	intent := bindingInfo.Intent[0]

	var ingressDenyRules []api.IngressDenyRule

	for _, resource := range intent.Spec.Intent.Resource {
		ingressRule := api.IngressDenyRule{}

		for _, cidrSet := range resource.FromCIDRSet {
			ingressRule.FromCIDRSet = append(ingressRule.FromCIDRSet, api.CIDRRule{
				Cidr: api.CIDR(cidrSet.CIDR),
			})
		}

		for _, toPort := range resource.ToPorts {
			var ports []api.PortProtocol
			for _, port := range toPort.Ports {
				ports = append(ports, api.PortProtocol{
					Port:     port.Port,
					Protocol: parseProtocol(port.Protocol),
				})
			}
			ingressRule.ToPorts = api.PortDenyRules{
				{
					Ports: ports,
				},
			}
		}

		ingressDenyRules = append(ingressDenyRules, ingressRule)
	}

	return ingressDenyRules
}

func parseProtocol(protocol string) api.L4Proto {
	// Convert protocol string to L4Proto type.
	switch strings.ToUpper(protocol) {
	case "TCP":
		return api.ProtoTCP
	case "UDP":
		return api.ProtoUDP
	case "ICMP":
		return api.ProtoICMP
	default:
		return api.ProtoTCP
	}
}

// ----------------------------------------
// -------- Apply & Update Policy  --------
// ----------------------------------------

// ApplyOrUpdatePolicy applies or updates the given policy.
func ApplyOrUpdatePolicy(ctx context.Context, c client.Client, policy client.Object, policyName string) error {
	// Update the policy if it already exists, otherwise create a new one.
	log := log.FromContext(ctx)

	var existingPolicy client.Object
	var policySpec interface{}

	switch p := policy.(type) {
	case *kubearmorv1.KubeArmorPolicy:
		existingPolicy = &kubearmorv1.KubeArmorPolicy{}
		policySpec = p.Spec
	case *kubearmorv1.KubeArmorHostPolicy:
		existingPolicy = &kubearmorv1.KubeArmorHostPolicy{}
		policySpec = p.Spec
	case *ciliumv2.CiliumNetworkPolicy:
		existingPolicy = &ciliumv2.CiliumNetworkPolicy{}
		policySpec = p.Spec
	default:
		return fmt.Errorf("unsupported policy type")
	}

	err := c.Get(ctx, types.NamespacedName{Name: policyName, Namespace: policy.GetNamespace()}, existingPolicy)
	if err != nil && !errors.IsNotFound(err) {
		// Other error handling
		log.Error(err, "Failed to get existing policy", "policy", policyName)
		return err
	}

	if errors.IsNotFound(err) {
		// Create a policy if it doesn't exist
		if err := c.Create(ctx, policy); err != nil {
			log.Error(err, "Failed to apply policy", "policy", policyName)
			return err
		}
		log.Info("Policy created", "Name", policyName)
	} else {
		// Update if policy already exists (compares specs only)
		existingSpec := reflect.ValueOf(existingPolicy).Elem().FieldByName("Spec").Interface()
		if !reflect.DeepEqual(policySpec, existingSpec) {
			reflect.ValueOf(existingPolicy).Elem().FieldByName("Spec").Set(reflect.ValueOf(policySpec))
			if err := c.Update(ctx, existingPolicy); err != nil {
				log.Error(err, "Failed to update policy", "policy", policyName)
				return err
			}
			log.Info("Policy updated", "Name", policyName)
		} else {
			log.Info("Policy unchanged", "Name", policyName)
		}
	}
	return nil
}

// ----------------------------------------
// ----------- Delete Policy  -------------
// ----------------------------------------

// DeletePolicy deletes a policy based on type, name, and namespace.
func DeletePolicy(ctx context.Context, c client.Client, policyType, name, namespace string) error {
	// Process the deletion request based on policy type.

	var policy client.Object
	log := log.FromContext(ctx)

	switch policyType {
	case "KubeArmorPolicy":
		policy = &kubearmorv1.KubeArmorPolicy{}
	case "KubeArmorHostPolicy":
		policy = &kubearmorv1.KubeArmorHostPolicy{}
	case "CiliumNetworkPolicy":
		policy = &ciliumv2.CiliumNetworkPolicy{}
	default:
		return fmt.Errorf("Unknown policy type: %s", policyType)
	}

	policy.SetName(name)
	policy.SetNamespace(namespace)

	if err := c.Delete(ctx, policy); client.IgnoreNotFound(err) != nil {
		log.Error(err, "Failed to delete policy", "Type", policyType, "Name", name, "Namespace", namespace)
		return err
	}
	return nil
}
*/
