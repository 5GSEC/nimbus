// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package utils

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	intentv1 "github.com/5GSEC/nimbus/api/v1"
	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/cilium/cilium/pkg/policy/api"
	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
)

// ---------------------------------------------------
// -------- Creation of Policy Specifications --------
// ---------------------------------------------------

// BuildKubeArmorPolicySpec creates a policy specification (either KubeArmorPolicy or KubeArmorHostPolicy)
// based on the provided SecurityIntent and the type of policy.
func BuildKubeArmorPolicySpec(ctx context.Context, intent *intentv1.SecurityIntent, policyType string) interface{} {
	log := log.FromContext(ctx)
	// Logging the creation of a KubeArmor policy.
	log.Info("Creating KubeArmorPolicy", "Name", intent.Name)

	matchLabels := convertToMapString(extractLabels(intent))

	// Convert extracted information into specific KubeArmor policy types.
	if policyType == "host" {

		return &kubearmorv1.KubeArmorHostPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      intent.Name,
				Namespace: intent.Namespace,
			},
			Spec: kubearmorv1.KubeArmorHostPolicySpec{
				NodeSelector: kubearmorv1.NodeSelectorType{
					MatchLabels: matchLabels,
				},
				Process:      convertToKubeArmorHostPolicyProcessType(extractProcessPolicy(intent)),
				File:         convertToKubeArmorHostPolicyFileType(extractFilePolicy(intent)),
				Capabilities: convertToKubeArmorHostPolicyCapabilitiesType(extractCapabilitiesPolicy(intent)),
				// TODO: To discuss
				//Network:      convertToKubeArmorHostPolicyNetworkType(extractNetworkPolicy(intent)),
				Action: kubearmorv1.ActionType(formatAction(intent.Spec.Intent.Action)),
			},
		}
	} else {
		return &kubearmorv1.KubeArmorPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      intent.Name,
				Namespace: intent.Namespace,
			},
			Spec: kubearmorv1.KubeArmorPolicySpec{
				Selector: kubearmorv1.SelectorType{
					MatchLabels: matchLabels,
				},
				Process:      convertToKubeArmorPolicyProcessType(extractProcessPolicy(intent)),
				File:         convertToKubeArmorPolicyFileType(extractFilePolicy(intent)),
				Capabilities: convertToKubeArmorPolicyCapabilitiesType(extractCapabilitiesPolicy(intent)),
				Network:      convertToKubeArmorPolicyNetworkType(extractNetworkPolicy(intent)),
				Action:       kubearmorv1.ActionType(formatAction(intent.Spec.Intent.Action)),
			},
		}
	}
}

// BuildCiliumNetworkPolicySpec creates a Cilium network policy specification based on the provided SecurityIntent.
func BuildCiliumNetworkPolicySpec(ctx context.Context, intent *intentv1.SecurityIntent) interface{} {
	// Logging the creation of a Cilium Network Policy.
	log := log.FromContext(ctx)
	log.Info("Creating CiliumNetworkPolicy", "Name", intent.Name)

	// Utilize utility functions to construct a Cilium network policy from the intent.
	endpointSelector := getEndpointSelector(intent)
	ingressDenyRules := getIngressDenyRules(intent)

	// Build and return a Cilium Network Policy based on the intent.
	return &ciliumv2.CiliumNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      intent.Name,
			Namespace: intent.Namespace,
		},
		Spec: &api.Rule{
			EndpointSelector: endpointSelector,
			IngressDeny:      ingressDenyRules,
		},
	}
}

// ----------------------------------------
// -------- Conversion Functions ----------
// ----------------------------------------

// Various conversion functions to transform data from the SecurityIntent into specific policy types.
// These functions handle different aspects like processing types, file types, network types, etc.,
// and convert them into the format required by KubeArmor and Cilium policies.

// convertToMapString converts a slice of interfaces to a map of string key-value pairs.
func convertToMapString(slice []interface{}) map[string]string {
	// Iterate through the slice, converting each item into a map and merging them into a single map.
	result := make(map[string]string)
	for _, item := range slice {
		for key, value := range item.(map[string]string) {
			result[key] = value
		}
	}
	return result
}

func convertToKubeArmorHostPolicyProcessType(slice []interface{}) kubearmorv1.ProcessType {
	var result kubearmorv1.ProcessType
	for _, item := range slice {
		str, ok := item.(string)
		if !ok {
			continue // or appropriate error handling
		}
		// The 'Pattern' field is of type string, so it can be assigned directly
		result.MatchPatterns = append(result.MatchPatterns, kubearmorv1.ProcessPatternType{
			Pattern: str,
		})
	}
	return result
}

func convertToKubeArmorHostPolicyFileType(slice []interface{}) kubearmorv1.FileType {
	var result kubearmorv1.FileType
	for _, item := range slice {
		result.MatchPaths = append(result.MatchPaths, kubearmorv1.FilePathType{
			Path: kubearmorv1.MatchPathType(item.(string)),
		})
	}
	return result
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

func convertToKubeArmorHostPolicyCapabilitiesType(slice []interface{}) kubearmorv1.HostCapabilitiesType {
	var result kubearmorv1.HostCapabilitiesType
	for _, item := range slice {
		str, ok := item.(string)
		if !ok {
			continue // or appropriate error handling
		}
		// Convert to MatchCapabilitiesStringType
		capability := kubearmorv1.MatchCapabilitiesStringType(str)
		result.MatchCapabilities = append(result.MatchCapabilities, kubearmorv1.MatchHostCapabilitiesType{
			Capability: capability,
		})
	}
	return result
}

func convertToKubeArmorPolicyProcessType(slice []interface{}) kubearmorv1.ProcessType {
	var result kubearmorv1.ProcessType
	for _, item := range slice {
		if str, ok := item.(string); ok {
			result.MatchPatterns = append(result.MatchPatterns, kubearmorv1.ProcessPatternType{
				Pattern: str,
			})
		}
	}
	return result
}

func convertToKubeArmorPolicyFileType(slice []interface{}) kubearmorv1.FileType {
	var result kubearmorv1.FileType
	for _, item := range slice {
		path, ok := item.(map[string]string)["Path"]
		if !ok {
			continue // or appropriate error handling
		}
		result.MatchPaths = append(result.MatchPaths, kubearmorv1.FilePathType{
			Path: kubearmorv1.MatchPathType(path),
		})
	}
	return result
}

func convertToKubeArmorPolicyCapabilitiesType(slice []interface{}) kubearmorv1.CapabilitiesType {
	var result kubearmorv1.CapabilitiesType
	//var types []kubearmorv1.MatchCapabilitiesType
	//var spec *kubearmorv1.KubeArmorPolicySpec
	for _, item := range slice {
		str, ok := item.(map[string]string)["Capability"]
		if !ok {
			continue // or appropriate error handling
		}
		result.MatchCapabilities = append(result.MatchCapabilities, kubearmorv1.MatchCapabilitiesType{
			Capability: kubearmorv1.MatchCapabilitiesStringType(str),
		})
	}
	if len(result.MatchCapabilities) == 0 {
		return kubearmorv1.CapabilitiesType{
			MatchCapabilities: []kubearmorv1.MatchCapabilitiesType{},
		}
	}
	return result
}

func convertToKubeArmorPolicyNetworkType(slice []interface{}) kubearmorv1.NetworkType {
	var result kubearmorv1.NetworkType
	for _, item := range slice {
		str, ok := item.(map[string]string)["Protocol"]
		if !ok {
			continue // or appropriate error handling
		}
		result.MatchProtocols = append(result.MatchProtocols, kubearmorv1.MatchNetworkProtocolType{
			Protocol: kubearmorv1.MatchNetworkProtocolStringType(str),
		})
	}
	return result
}

// --------------------------------------
// -------- Utility Functions  ----------
// --------------------------------------

// GetPolicyType returns the type of policy as a string based on whether it's a host policy.
func GetPolicyType(isHost bool) string {
	if isHost {
		return "KubeArmorHostPolicy"
	}
	return "KubeArmorPolicy"
}

// IsHostPolicy determines if the given SecurityIntent is a Host Policy.
func IsHostPolicy(intent *intentv1.SecurityIntent) bool {
	// Check for specific labels in the CEL field of the intent to determine if it's a host policy.
	for _, cel := range intent.Spec.Selector.CEL {
		if strings.Contains(cel, "kubernetes.io") {
			return true
		}
	}
	return false
}

// cleanLabelKey cleans up the label key to remove unnecessary prefixes.
func cleanLabelKey(key string) string {
	// Remove specific prefixes from the label key, if present.
	if strings.HasPrefix(key, "object.metadata.labels.") {
		return strings.TrimPrefix(key, "object.metadata.labels.")
	}
	return key
}

// parseCELExpression parses a CEL expression and extracts labels as key-value pairs.
func parseCELExpression(expression string) map[string]string {
	// Process the CEL expression and convert it into a map of labels.
	parsedLabels := make(map[string]string)

	expression = strings.TrimSpace(expression)
	expressions := strings.Split(expression, " && ")

	for _, expr := range expressions {
		parts := strings.Split(expr, " == ")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			key = strings.TrimPrefix(key, "object.metadata.labels['")
			key = strings.TrimSuffix(key, "']")
			value = strings.Trim(value, "'")

			parsedLabels[key] = value
		}
	}

	return parsedLabels
}

// extractLabels(): Extracts and returns labels from a CEL expression and Match fields
func extractLabels(intent *intentv1.SecurityIntent) []interface{} {
	labels := make([]interface{}, 0)

	isHost := IsHostPolicy(intent)

	// Extract labels from Match.Any
	for _, filter := range intent.Spec.Selector.Match.Any {
		for key, val := range filter.Resources.MatchLabels {
			if strings.HasPrefix(key, "object.metadata.labels.") {
				cleanKey := cleanLabelKey(key)
				labels = append(labels, map[string]string{cleanKey: val})
			}
		}
	}

	// Extract labels from Match.All
	for _, filter := range intent.Spec.Selector.Match.All {
		for key, val := range filter.Resources.MatchLabels {
			if strings.HasPrefix(key, "object.metadata.labels.") {
				cleanKey := cleanLabelKey(key)
				labels = append(labels, map[string]string{cleanKey: val})
			}
		}
	}

	// Extract labels from CEL expressions
	if isHost {
		// Special handling for KubeArmorHostPolicy
		for _, cel := range intent.Spec.Selector.CEL {
			if strings.Contains(cel, "kubernetes.io/") {
				parts := strings.Split(cel, " == ")
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					key = strings.TrimPrefix(key, "object.metadata.labels['")
					key = strings.TrimSuffix(key, "']")
					value := strings.Trim(parts[1], "'")
					cleanKey := cleanLabelKey(key)
					labels = append(labels, map[string]string{cleanKey: value})
				}
			}
		}
	} else {
		// General handling for other policies
		for _, cel := range intent.Spec.Selector.CEL {
			if strings.HasPrefix(cel, "object.metadata.labels.") {
				parsedLabels := parseCELExpression(cel)
				for key, val := range parsedLabels {
					cleanKey := cleanLabelKey(key)
					labels = append(labels, map[string]string{cleanKey: val})
				}
			}
		}
	}

	return labels
}

func extractProcessPolicy(intent *intentv1.SecurityIntent) []interface{} {
	var matchPatterns []interface{}
	for _, resource := range intent.Spec.Intent.Resource {
		if resource.Key == "commands" {
			for _, cmd := range resource.Val {
				matchPatterns = append(matchPatterns, map[string]string{"Pattern": cmd})
			}
		}
	}
	return matchPatterns
}

func extractFilePolicy(intent *intentv1.SecurityIntent) []interface{} {
	var matchPaths []interface{}
	for _, resource := range intent.Spec.Intent.Resource {
		if resource.Key == "paths" {
			for _, path := range resource.Val {
				matchPaths = append(matchPaths, map[string]string{"Path": path})
			}
		}
	}
	return matchPaths
}

func extractCapabilitiesPolicy(intent *intentv1.SecurityIntent) []interface{} {
	var matchCapabilities []interface{}
	for _, resource := range intent.Spec.Intent.Resource {
		if resource.Key == "capabilities" {
			for _, capability := range resource.Val {
				matchCapabilities = append(matchCapabilities, map[string]string{"Capability": capability})
			}
		}
	}
	return matchCapabilities
}

// extractNetworkPolicy() - Extracts network policy from SecurityIntent and returns it as a slice of interface{}
func extractNetworkPolicy(intent *intentv1.SecurityIntent) []interface{} {
	var matchNetworkProtocols []interface{}

	for _, resource := range intent.Spec.Intent.Resource {
		if resource.Key == "protocols" {
			for _, protocol := range resource.Val {
				protocolMap := map[string]string{"Protocol": protocol}
				matchNetworkProtocols = append(matchNetworkProtocols, protocolMap)
			}
		}
	}

	return matchNetworkProtocols
}

// getEndpointSelector creates an endpoint selector from the SecurityIntent.
func getEndpointSelector(intent *intentv1.SecurityIntent) api.EndpointSelector {
	// Create an Endpoint Selector based on matched labels extracted from the intent.
	matchLabels := make(map[string]string)

	// Matching labels to a "Match Any" filter
	for _, filter := range intent.Spec.Selector.Match.Any {
		for key, val := range filter.Resources.MatchLabels {
			matchLabels[key] = val
		}
	}

	// Matching labels that fit the "Match All" filter
	for _, filter := range intent.Spec.Selector.Match.All {
		for key, val := range filter.Resources.MatchLabels {
			matchLabels[key] = val
		}
	}

	// Create an Endpoint Selector based on matched labels
	return api.NewESFromMatchRequirements(matchLabels, nil)
}

// getIngressDenyRules generates ingress deny rules from SecurityIntent.
func getIngressDenyRules(intent *intentv1.SecurityIntent) []api.IngressDenyRule {
	// Process the intent to create ingress deny rules.
	var ingressDenyRules []api.IngressDenyRule

	for _, resource := range intent.Spec.Intent.Resource {
		if resource.Key == "ingress" {
			for _, val := range resource.Val {
				cidr, port, protocol := splitCIDRAndPort(val)

				ingressRule := api.IngressDenyRule{
					ToPorts: api.PortDenyRules{
						{
							Ports: []api.PortProtocol{
								{
									Port:     port,
									Protocol: parseProtocol(protocol),
								},
							},
						},
					},
				}

				if cidr != "" {
					ingressRule.FromCIDRSet = []api.CIDRRule{{Cidr: api.CIDR(cidr)}}
				}

				ingressDenyRules = append(ingressDenyRules, ingressRule)
			}
		}
	}

	return ingressDenyRules
}

// splitCIDRAndPort separates CIDR, port, and protocol information from a combined string.
func splitCIDRAndPort(cidrAndPort string) (string, string, string) {
	// Split the string into CIDR, port, and protocol components, with handling for different formats.

	// Default protocol is TCP
	defaultProtocol := "TCP"

	// Separate strings based on '-'
	split := strings.Split(cidrAndPort, "-")

	// If there are three separate elements, return the CIDR, port, and protocol separately
	if len(split) == 3 {
		return split[0], split[1], split[2]
	}

	// If there are two separate elements, return the CIDR, port, and default protocol
	if len(split) == 2 {
		return split[0], split[1], defaultProtocol
	}

	// If there is only one element, return the CIDR, empty port, and default protocol
	return cidrAndPort, "", defaultProtocol
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

func formatAction(action string) string {
	// Convert action string to a specific format.
	switch strings.ToLower(action) {
	case "block":
		return "Block"
	case "audit":
		return "Audit"
	case "allow":
		return "Allow"
	default:
		return action
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
		return fmt.Errorf("unknown policy type: %s", policyType)
	}

	policy.SetName(name)
	policy.SetNamespace(namespace)

	if err := c.Delete(ctx, policy); client.IgnoreNotFound(err) != nil {
		log.Error(err, "Failed to delete policy", "Type", policyType, "Name", name, "Namespace", namespace)
		return err
	}
	return nil
}

// containsString checks if a string is included in a slice.
func ContainsString(slice []string, s string) bool {
	// Iterate through the slice to check if the string exists.

	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// removeString removes a specific string from a slice.
func RemoveString(slice []string, s string) []string {
	// Create a new slice excluding the specified string.
	result := []string{}
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}
