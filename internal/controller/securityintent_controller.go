/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"reflect"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	intentv1 "git.cclab-inu.com/b0m313/nimbus/api/v1"
	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/cilium/cilium/pkg/policy/api"
	kubearmorhostpolicyv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorHostPolicy/api/security.kubearmor.com/v1"
	kubearmorpolicyv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorPolicy/api/security.kubearmor.com/v1"
)

// SecurityIntentReconciler reconciles a SecurityIntent object
type SecurityIntentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintents/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecurityIntent object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcil

const securityIntentFinalizer = "finalizer.securityintent.intent.security.nimbus.com"

// Reconcile attempts to bring the current state of the cluster closer to the desired state.
func (r *SecurityIntentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Retrieve the SecurityIntent resource
	intent := &intentv1.SecurityIntent{}
	err := r.Get(ctx, req.NamespacedName, intent)

	// Handle cases where the resource is deleted or an error occurred during retrieval
	if err != nil {
		if errors.IsNotFound(err) {
			// Log and exit when the resource is deleted
			log.Info("SecurityIntent resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Other errors are returned after logging
		log.Error(err, "Failed to get SecurityIntent")
		return ctrl.Result{}, err
	}

	// Gets a SecurityIntent object
	intent, err = GeneralController(ctx, r.Client, req.NamespacedName.Name, req.NamespacedName.Namespace)
	if err != nil {
		log.Error(err, "Failed fetching SecurityIntent")
		return ctrl.Result{}, err
	}

	// Create and enforce security intent-based policies as actual policies
	if err = PolicyController(ctx, intent, r.Client); err != nil {
		log.Error(err, "Failed applying policy")
		return ctrl.Result{}, err
	}

	// Finalizer
	if intent.ObjectMeta.DeletionTimestamp.IsZero() {
		// If the SecurityIntent is not being deleted
		if !containsString(intent.ObjectMeta.Finalizers, securityIntentFinalizer) {
			intent.ObjectMeta.Finalizers = append(intent.ObjectMeta.Finalizers, securityIntentFinalizer)
			if err := r.Update(ctx, intent); err != nil {
				return ctrl.Result{}, err
			}

		}
	} else {
		// If the SecurityIntent is being deleted
		if containsString(intent.ObjectMeta.Finalizers, securityIntentFinalizer) {
			// Related policy deletion logic
			if err := deleteRelatedPolicies(ctx, r.Client, intent); err != nil {
				return ctrl.Result{}, err
			}

			// Remove Finalizer
			intent.ObjectMeta.Finalizers = removeString(intent.ObjectMeta.Finalizers, securityIntentFinalizer)
			if err := r.Update(ctx, intent); err != nil {
				return ctrl.Result{}, err
			}

		}
		return ctrl.Result{}, nil
	}

	log.Info("Successfully reconciled SecurityIntent", "intent", intent.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecurityIntentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&intentv1.SecurityIntent{}).
		Complete(r)
}

// ------------------------------
// ---- General Controller ------
// ------------------------------

// GeneralController() : Detect and process SecurityIntent objects
func GeneralController(ctx context.Context, client client.Client, intentName string, namespace string) (*intentv1.SecurityIntent, error) {
	log := log.FromContext(ctx)
	log.Info("Searching for SecurityIntent", "Name", intentName, "Namespace", namespace)

	// Search for a SecurityIntent object
	var intent intentv1.SecurityIntent
	if err := client.Get(ctx, types.NamespacedName{Name: intentName, Namespace: namespace}, &intent); err != nil {
		log.Error(err, "Failed searching SecurityIntent")
		return nil, err
	}

	log.Info("Found SecurityIntent", "Name", intent.Name, "Namespace", namespace)
	return &intent, nil
}

// extractLabels(): Extracts and returns labels from a CEL expression and Match fields
func extractLabels(intent *intentv1.SecurityIntent, isHostPolicy bool) map[string]string {
	labels := make(map[string]string)

	// Extract labels from Match.Any
	for _, filter := range intent.Spec.Selector.Match.Any {
		for key, val := range filter.Resources.MatchLabels {
			if strings.HasPrefix(key, "object.metadata.labels.") {
				cleanKey := cleanLabelKey(key)
				labels[cleanKey] = val
			}
		}
	}

	// Extract labels from Match.All
	for _, filter := range intent.Spec.Selector.Match.All {
		for key, val := range filter.Resources.MatchLabels {
			if strings.HasPrefix(key, "object.metadata.labels.") {
				cleanKey := cleanLabelKey(key)
				labels[cleanKey] = val
			}
		}
	}

	// Extract labels from CEL expressions
	if isHostPolicy {
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
					labels[cleanKey] = value
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
					labels[cleanKey] = val
				}
			}
		}
	}

	return labels
}

// cleanLabelKey(): Cleans up the label key to remove unnecessary prefixes
func cleanLabelKey(key string) string {
	if strings.HasPrefix(key, "object.metadata.labels.") {
		return strings.TrimPrefix(key, "object.metadata.labels.")
	}
	return key
}

// parseCELExpression: Parses a CEL expression and extracts labels
func parseCELExpression(expression string) map[string]string {
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

// ------------------------------
// ---- Policy Controller -------
// ------------------------------

// PolicyController() : Creates, applies, modifies, and deletes actual policies based on the policies defined in the SecurityIntent.
func PolicyController(ctx context.Context, intent *intentv1.SecurityIntent, c client.Client) error {
	log := log.FromContext(ctx)
	log.Info("Processing policy", "Name", intent.Name, "Type", intent.Spec.Intent.Type)

	switch intent.Spec.Intent.Type {
	case "system":
		if err := handleSystemPolicy(ctx, intent, c); err != nil {
			log.Error(err, "Failed handling system policy")
			return err
		}
	case "network":
		if err := handleNetworkPolicy(ctx, intent, c); err != nil {
			log.Error(err, "Failed handling network policy")
			return err
		}
		if err := handleKubeArmorNetworkPolicy(ctx, intent, c); err != nil {
			log.Error(err, "Failed handling KubeArmor network policy")
			return err
		}
	default:
		log.Info("Encountered unknown policy type", "type", intent.Spec.Intent.Type)
	}
	log.Info("Policy processing completed", "Name", intent.Name)
	return nil
}

// ---------------------------------------
// ------ System Policy  --------
// ---------------------------------------

// handleSystemPolicy(): Handle system policy
func handleSystemPolicy(ctx context.Context, intent *intentv1.SecurityIntent, c client.Client) error {
	log := log.FromContext(ctx)

	// Create and apply the appropriate policy based on whether it's a HostPolicy or not
	if isHostPolicy(intent) {
		hostPolicy := createKubeArmorHostPolicy(ctx, intent, isHostPolicy(intent))
		if err := applyKubeArmorHostPolicy(ctx, c, hostPolicy); err != nil {
			return err
		}
		log.Info("Applied System Policy")
	} else {
		armorPolicy := createKubeArmorPolicy(ctx, intent, isHostPolicy(intent))
		if err := applyKubeArmorPolicy(ctx, c, armorPolicy); err != nil {
			return err
		}
		log.Info("Applied System Policy")
	}
	return nil
}

// ---------------------------------------
// ------ Generate System Policy  --------
// ---------------------------------------

// createKubeArmorHostPolicy(): Creates a KubeArmorHostPolicy object based on the given SecurityIntent
func createKubeArmorHostPolicy(ctx context.Context, intent *intentv1.SecurityIntent, isHost bool) *kubearmorhostpolicyv1.KubeArmorHostPolicy {
	log := log.FromContext(ctx)
	log.Info("Creating KubeArmorHostPolicy", "Name", intent.Name)

	process, file, capabilities, network := buildPolicySpec(intent)
	if process == nil {
		process = &kubearmorhostpolicyv1.ProcessType{}
	}
	if file == nil {
		file = &kubearmorhostpolicyv1.FileType{}
	}
	if capabilities == nil {
		capabilities = &kubearmorhostpolicyv1.CapabilitiesType{}
	}
	if network == nil {
		network = &kubearmorhostpolicyv1.NetworkType{}
	}

	matchLabels := extractLabels(intent, isHost)

	spec := kubearmorhostpolicyv1.KubeArmorHostPolicySpec{
		NodeSelector: kubearmorhostpolicyv1.NodeSelectorType{
			MatchLabels: matchLabels,
		},
		Process:      *process,
		File:         *file,
		Capabilities: *capabilities,
		Network:      *network,
	}

	log.Info("KubeArmorHostPolicy created", "Name", intent.Name)
	return &kubearmorhostpolicyv1.KubeArmorHostPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      intent.Name,
			Namespace: intent.Namespace,
		},
		Spec: spec,
	}
}

// createKubeArmorPolicy() - Creates a KubeArmor policy based on SecurityIntent
func createKubeArmorPolicy(ctx context.Context, intent *intentv1.SecurityIntent, isHost bool) *kubearmorpolicyv1.KubeArmorPolicy {
	log := log.FromContext(ctx)
	log.Info("Creating KubeArmorPolicy", "Name", intent.Name)

	// Extract label match conditions from SecurityIntent
	matchLabels := extractLabels(intent, isHost)

	// Constructing KubeArmorPolicySpec based on the intent
	armorPolicySpec := kubearmorpolicyv1.KubeArmorPolicySpec{
		Selector: kubearmorpolicyv1.SelectorType{
			MatchLabels: matchLabels,
		},
		Process:      extractProcessPolicy(intent),
		File:         extractFilePolicy(intent),
		Network:      extractNetworkPolicy(intent),
		Capabilities: extractCapabilitiesPolicy(intent),
		Action:       formatActionForArmorPolicy(intent.Spec.Intent.Action),
	}

	// Creating the KubeArmorPolicy object
	log.Info("KubeArmorPolicy created", "Name", intent.Name)
	return &kubearmorpolicyv1.KubeArmorPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      intent.Name,
			Namespace: intent.Namespace,
		},
		Spec: armorPolicySpec,
	}
}

// buildPolicySpec(): Configures the Process, File, Capabilities, Network fields in the given SecurityIntent
func buildPolicySpec(intent *intentv1.SecurityIntent) (*kubearmorhostpolicyv1.ProcessType, *kubearmorhostpolicyv1.FileType, *kubearmorhostpolicyv1.CapabilitiesType, *kubearmorhostpolicyv1.NetworkType) {
	var process *kubearmorhostpolicyv1.ProcessType
	var file *kubearmorhostpolicyv1.FileType
	var capabilities *kubearmorhostpolicyv1.CapabilitiesType
	var network *kubearmorhostpolicyv1.NetworkType

	// Configure the Process and File fields
	if len(intent.Spec.Intent.Resource) > 0 {
		for _, resource := range intent.Spec.Intent.Resource {
			if resource.Key == "paths" {
				for _, val := range resource.Val {
					if isExecutablePath(val) {
						// Executable path, add to process match patterns
						if process == nil {
							process = &kubearmorhostpolicyv1.ProcessType{
								MatchPatterns: []kubearmorhostpolicyv1.ProcessPatternType{},
								Action:        kubearmorhostpolicyv1.ActionType(intent.Spec.Intent.Action),
							}
						}
						process.MatchPatterns = append(process.MatchPatterns, kubearmorhostpolicyv1.ProcessPatternType{
							Pattern: val,
						})
					} else {
						// Non-executable path, add to file match paths
						if file == nil {
							file = &kubearmorhostpolicyv1.FileType{
								MatchPaths: []kubearmorhostpolicyv1.FilePathType{},
								Action:     kubearmorhostpolicyv1.ActionType(intent.Spec.Intent.Action),
							}
						}
						file.MatchPaths = append(file.MatchPaths, kubearmorhostpolicyv1.FilePathType{
							Path: kubearmorhostpolicyv1.MatchPathType(val),
						})
					}
				}
			}
		}
	}

	// Set default values for capabilities and network if they are nil
	if capabilities == nil {
		capabilities = &kubearmorhostpolicyv1.CapabilitiesType{
			MatchCapabilities: []kubearmorhostpolicyv1.MatchCapabilitiesType{},
		}
	}
	if network == nil {
		network = &kubearmorhostpolicyv1.NetworkType{
			MatchProtocols: []kubearmorhostpolicyv1.MatchNetworkProtocolType{},
		}
	}

	return process, file, capabilities, network
}

// isHostPolicy(): Determines if the given SecurityIntent is a Host Policy
func isHostPolicy(intent *intentv1.SecurityIntent) bool {
	// Check for the kubernetes.io label in the CEL field
	for _, cel := range intent.Spec.Selector.CEL {
		if strings.Contains(cel, "kubernetes.io") {
			return true
		}
	}
	return false
}

// isExecutablePath: Determines if a given path is likely an executable
func isExecutablePath(path string) bool {
	// Define common executable paths
	executablePaths := []string{"/bin", "/usr/bin", "/sbin", "/usr/sbin", "/usr/local/bin"}

	// Check if the path starts with any of the common executable paths
	for _, prefix := range executablePaths {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// formatPatternWithValcel() : Form the entire pattern based on the 'valcel' pattern
func formatPatternWithValcel(valcel, command string) string {
	if strings.HasPrefix(valcel, "pattern: ") {
		pattern := strings.TrimPrefix(valcel, "pattern: ")
		// e.g. 'pattern: /**/{command}'
		return strings.Replace(pattern, "{command}", command, -1)
	}
	return command
}

// extractMatchPaths(): Extract file paths from the resource
func extractMatchPaths(resources []intentv1.Resource) []kubearmorhostpolicyv1.FilePathType {
	var matchPaths []kubearmorhostpolicyv1.FilePathType
	for _, resource := range resources {
		if resource.Key == "paths" {
			for _, path := range resource.Val {
				matchPaths = append(matchPaths, kubearmorhostpolicyv1.FilePathType{
					Path: kubearmorhostpolicyv1.MatchPathType(path),
				})
			}
		}
	}
	return matchPaths
}

// extractProcessPolicy() - Extracts process policy from SecurityIntent
func extractProcessPolicy(intent *intentv1.SecurityIntent) kubearmorpolicyv1.ProcessType {
	var processPolicy kubearmorpolicyv1.ProcessType

	for _, resource := range intent.Spec.Intent.Resource {
		if resource.Key == "commands" {
			for _, cmd := range resource.Val {
				processPolicy.MatchPatterns = append(processPolicy.MatchPatterns, kubearmorpolicyv1.ProcessPatternType{
					Pattern: cmd,
				})
			}
		}
	}

	return processPolicy
}

// extractFilePolicy() - Extracts file policy from SecurityIntent
func extractFilePolicy(intent *intentv1.SecurityIntent) kubearmorpolicyv1.FileType {
	var filePolicy kubearmorpolicyv1.FileType

	for _, resource := range intent.Spec.Intent.Resource {
		if resource.Key == "paths" {
			for _, path := range resource.Val {
				filePolicy.MatchPaths = append(filePolicy.MatchPaths, kubearmorpolicyv1.FilePathType{
					Path: kubearmorpolicyv1.MatchPathType(path), // 올바른 타입으로 수정
				})
			}
		}
	}

	return filePolicy
}

// extractCapabilitiesPolicy() - Extracts capabilities policy from SecurityIntent
func extractCapabilitiesPolicy(intent *intentv1.SecurityIntent) kubearmorpolicyv1.CapabilitiesType {
	var capabilitiesPolicy kubearmorpolicyv1.CapabilitiesType

	// Check for capabilities in the SecurityIntent's resources
	for _, resource := range intent.Spec.Intent.Resource {
		if resource.Key == "capabilities" {
			for _, capability := range resource.Val {
				// Add each capability to the MatchCapabilities slice
				capabilitiesPolicy.MatchCapabilities = append(capabilitiesPolicy.MatchCapabilities, kubearmorpolicyv1.MatchCapabilitiesType{
					Capability: kubearmorpolicyv1.MatchCapabilitiesStringType(capability),
				})
			}
		}
	}

	// Ensure that MatchCapabilities is always initialized, even if it's empty
	if len(capabilitiesPolicy.MatchCapabilities) == 0 {
		capabilitiesPolicy.MatchCapabilities = []kubearmorpolicyv1.MatchCapabilitiesType{}
	}

	return capabilitiesPolicy
}

// ---------------------------------------
// ------- Apply System Policy  ----------
// ---------------------------------------

func applyKubeArmorPolicy(ctx context.Context, c client.Client, policy *kubearmorpolicyv1.KubeArmorPolicy) error {
	return applyOrUpdatePolicy(ctx, c, policy, policy.Name)
}

func applyKubeArmorHostPolicy(ctx context.Context, c client.Client, policy *kubearmorhostpolicyv1.KubeArmorHostPolicy) error {
	return applyOrUpdatePolicy(ctx, c, policy, policy.Name)
}

// -------------------------------
// ------ Network Policy  --------
// -------------------------------

// handleNetworkPolicy(): Handle Cilium network policy
func handleNetworkPolicy(ctx context.Context, intent *intentv1.SecurityIntent, c client.Client) error {
	log := log.FromContext(ctx)

	ciliumPolicy := createCiliumNetworkPolicy(ctx, intent)
	if err := applyCiliumNetworkPolicy(ctx, c, ciliumPolicy); err != nil {
		log.Error(err, "Failed applying CiliumNetworkPolicy", "policy", ciliumPolicy)
		return err
	}
	log.Info("Applied CiliumNetworkPolicy", "policy", ciliumPolicy)

	return nil // Returns nil on success
}

// handleKubeArmorNetworkPolicy(): Handle KubeArmor network policy
func handleKubeArmorNetworkPolicy(ctx context.Context, intent *intentv1.SecurityIntent, c client.Client) error {
	log := log.FromContext(ctx)

	// Only create KubeArmor Network Policy if intent type is 'network'
	if intent.Spec.Intent.Type == "network" {
		armorPolicy := createKubeArmorNetworkPolicy(ctx, intent, isHostPolicy(intent))
		if err := applyKubeArmorNetworkPolicy(ctx, c, armorPolicy); err != nil {
			log.Error(err, "Failed applying KubeArmorPolicy", "policy", armorPolicy.Name)
			return err
		}
		log.Info("Applied KubeArmorPolicy", "policy", armorPolicy)
	}

	return nil
}

// ----------------------------------------
// ------ Generate Network Policy  --------
// ----------------------------------------

// createKubeArmorNetworkPolicy(): Creates a KubeArmor policy for network-related intents
func createKubeArmorNetworkPolicy(ctx context.Context, intent *intentv1.SecurityIntent, isHost bool) *kubearmorpolicyv1.KubeArmorPolicy {
	log := log.FromContext(ctx)
	log.Info("Creating KubearmorPolicy", "Name", intent.Name)

	matchLabels := extractLabels(intent, isHost)

	armorPolicySpec := kubearmorpolicyv1.KubeArmorPolicySpec{
		Selector: kubearmorpolicyv1.SelectorType{
			MatchLabels: matchLabels,
		},
		Network: extractNetworkPolicy(intent),
		Action:  formatActionForArmorPolicy(intent.Spec.Intent.Action),
		Capabilities: kubearmorpolicyv1.CapabilitiesType{ // 추가된 부분
			MatchCapabilities: []kubearmorpolicyv1.MatchCapabilitiesType{},
		},
	}

	return &kubearmorpolicyv1.KubeArmorPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      intent.Name,
			Namespace: intent.Namespace,
		},
		Spec: armorPolicySpec,
	}
}

// createCiliumNetworkPolicy() : Create a CiliumNetworkPolicy
func createCiliumNetworkPolicy(ctx context.Context, intent *intentv1.SecurityIntent) *ciliumv2.CiliumNetworkPolicy {
	log := log.FromContext(ctx)
	log.Info("Creating CiliumNetworkPolicy", "Name", intent.Name)

	// Create a Endpoint Selector field
	endpointSelector := getEndpointSelector(intent)
	// Create a Ingress Deny Rule
	ingressDenyRules := getIngressDenyRules(intent)

	//egressRules := getEgressRules(intent)

	// Initializing a CiliumNetworkPolicy Object
	log.Info("CiliumNetworkPolicy created", "Name", intent.Name)
	return &ciliumv2.CiliumNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      intent.Name,      // Policy name
			Namespace: intent.Namespace, // Policy namespace
		},
		Spec: &api.Rule{
			EndpointSelector: endpointSelector, // endpoint selector
			IngressDeny:      ingressDenyRules, // Ingress Deny Rule
			//Egress:           egressRules,
		},
	}
}

// extractNetworkPolicy() - Extracts network policy from SecurityIntent
func extractNetworkPolicy(intent *intentv1.SecurityIntent) kubearmorpolicyv1.NetworkType {
	var networkPolicy kubearmorpolicyv1.NetworkType

	for _, resource := range intent.Spec.Intent.Resource {
		if resource.Key == "protocols" {
			for _, protocol := range resource.Val {
				networkPolicy.MatchProtocols = append(networkPolicy.MatchProtocols, kubearmorpolicyv1.MatchNetworkProtocolType{
					Protocol: kubearmorpolicyv1.MatchNetworkProtocolStringType(protocol),
				})
			}
		}
	}

	// Ensure that MatchProtocols is always initialized, even if it's empty
	if networkPolicy.MatchProtocols == nil {
		networkPolicy.MatchProtocols = []kubearmorpolicyv1.MatchNetworkProtocolType{}
	}

	return networkPolicy
}

// getEndpointSelector(): Creates an endpoint selector from the SecurityIntent
func getEndpointSelector(intent *intentv1.SecurityIntent) api.EndpointSelector {
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

// getIngressDenyRules(): Generate ingress deny rules from SecurityIntent
func getIngressDenyRules(intent *intentv1.SecurityIntent) []api.IngressDenyRule {
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

// splitCIDRAndPort(): Separates CIDR, port, and protocol information
func splitCIDRAndPort(cidrAndPort string) (string, string, string) {
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

// ---------------------------------------
// ------- Apply Network Policy  ---------
// ---------------------------------------

func applyCiliumNetworkPolicy(ctx context.Context, c client.Client, policy *ciliumv2.CiliumNetworkPolicy) error {
	return applyOrUpdatePolicy(ctx, c, policy, policy.Name)
}

func applyKubeArmorNetworkPolicy(ctx context.Context, c client.Client, policy *kubearmorpolicyv1.KubeArmorPolicy) error {
	return applyOrUpdatePolicy(ctx, c, policy, policy.Name)
}

// ------------------------------
// ------ Apply Policy  --------
// ------------------------------

// applyOrUpdatePolicy: Applies or updates the given policy
// Update the policy if it already exists, otherwise create a new one.
func applyOrUpdatePolicy(ctx context.Context, c client.Client, policy client.Object, policyName string) error {
	log := log.FromContext(ctx)

	var existingPolicy client.Object
	var policySpec interface{}

	switch p := policy.(type) {
	case *kubearmorpolicyv1.KubeArmorPolicy:
		existingPolicy = &kubearmorpolicyv1.KubeArmorPolicy{}
		policySpec = p.Spec
	case *kubearmorhostpolicyv1.KubeArmorHostPolicy:
		existingPolicy = &kubearmorhostpolicyv1.KubeArmorHostPolicy{}
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

// ------------------------------
// ------ Delete Policy  --------
// ------------------------------

// deleteRelatedPolicies: Delete the associated policies based on the SecurityIntent
func deleteRelatedPolicies(ctx context.Context, c client.Client, intent *intentv1.SecurityIntent) error {
	log := log.FromContext(ctx)
	log.Info("Deleting related policies", "Name", intent.Name)

	switch intent.Spec.Intent.Type {
	case "system":
		if isHostPolicy(intent) {
			// Delete KubeArmorHostPolicy
			if err := deleteKubeArmorHostPolicy(ctx, c, intent.Name, intent.Namespace); err != nil {
				return err
			}
		} else {
			// Delete KubeArmorPolicy
			if err := deleteKubeArmorPolicy(ctx, c, intent.Name, intent.Namespace); err != nil {
				return err
			}
		}

	case "network":
		// Delete CiliumNetworkPolicy
		if err := deleteCiliumNetworkPolicy(ctx, c, intent.Name, intent.Namespace); err != nil {
			return err
		}
		// Delete the protocol-specific KubeArmorPolicy if it exists
		if containsProtocolResource(intent) {
			if err := deleteKubeArmorPolicy(ctx, c, intent.Name, intent.Namespace); err != nil {
				return err
			}
		}
	}

	log.Info("Related policies deleted successfully", "Name", intent.Name)
	return nil
}

// deleteKubeArmorPolicy: Delete KubeArmorPolicy
func deleteKubeArmorPolicy(ctx context.Context, c client.Client, name, namespace string) error {
	log := log.FromContext(ctx)

	armorPolicy := &kubearmorpolicyv1.KubeArmorPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := c.Delete(ctx, armorPolicy); client.IgnoreNotFound(err) != nil {
		log.Error(err, "Failed to delete KubeArmorPolicy", "Name", name, "Namespace", namespace)
		return err
	}
	return nil
}

// deleteKubeArmorHostPolicy: Delete KubeArmorHostPolicy
func deleteKubeArmorHostPolicy(ctx context.Context, c client.Client, name, namespace string) error {
	log := log.FromContext(ctx)

	hostPolicy := &kubearmorhostpolicyv1.KubeArmorHostPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := c.Delete(ctx, hostPolicy); client.IgnoreNotFound(err) != nil {
		log.Error(err, "Failed to delete KubeArmorHostPolicy", "Name", name, "Namespace", namespace)
		return err
	}
	return nil
}

// deleteCiliumNetworkPolicy: Delete CiliumNetworkPolicy
func deleteCiliumNetworkPolicy(ctx context.Context, c client.Client, name, namespace string) error {
	log := log.FromContext(ctx)

	ciliumPolicy := &ciliumv2.CiliumNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := c.Delete(ctx, ciliumPolicy); client.IgnoreNotFound(err) != nil {
		log.Error(err, "Failed to delete CiliumNetworkPolicy", "Name", name, "Namespace", namespace)
		return err
	}
	return nil
}

// -----------------------
// ------- etc  ----------
// -----------------------

// formatAction(): Convert action value to a valid format
func formatActionForHostPolicy(action string) kubearmorhostpolicyv1.ActionType {
	switch strings.ToLower(action) {
	case "block":
		return kubearmorhostpolicyv1.ActionType("Block")
	case "audit":
		return kubearmorhostpolicyv1.ActionType("Audit")
	case "allow":
		return kubearmorhostpolicyv1.ActionType("Allow")
	default:
		return kubearmorhostpolicyv1.ActionType(action)
	}
}

// formatActionForArmorPolicy(): Convert action value to kubearmorpolicyv1.ActionType
func formatActionForArmorPolicy(action string) kubearmorpolicyv1.ActionType {
	switch strings.ToLower(action) {
	case "block":
		return kubearmorpolicyv1.ActionType("Block")
	case "audit":
		return kubearmorpolicyv1.ActionType("Audit")
	case "allow":
		return kubearmorpolicyv1.ActionType("Allow")
	default:
		return kubearmorpolicyv1.ActionType(action)
	}
}

func parseProtocol(protocol string) api.L4Proto {
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

// containsProtocolResource: Checking for the presence of protocol resources in SecurityIntent
func containsProtocolResource(intent *intentv1.SecurityIntent) bool {
	for _, resource := range intent.Spec.Intent.Resource {
		if resource.Key == "protocols" {
			return true
		}
	}
	return false
}

// containsString: Checking if a string is included in a slice
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// removeString: Remove a specific string from a slice
func removeString(slice []string, s string) []string {
	result := []string{}
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}
