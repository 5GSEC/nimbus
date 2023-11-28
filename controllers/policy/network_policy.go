package policy

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	intentv1 "github.com/5GSEC/nimbus/api/v1"
	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	kubearmorpolicyv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorPolicy/api/security.kubearmor.com/v1"

	utils "github.com/5GSEC/nimbus/controllers/utils"
)

// NetworkPolicyController struct to handle network policies.
type NetworkPolicyController struct {
	Client client.Client   // Client to interact with Kubernetes API.
	Scheme *runtime.Scheme // Scheme defines the runtime scheme of the Kubernetes objects.
}

// NewNetworkPolicyController creates a new instance of NetworkPolicyController.
func NewNetworkPolicyController(client client.Client, scheme *runtime.Scheme) *NetworkPolicyController {
	return &NetworkPolicyController{
		Client: client,
		Scheme: scheme,
	}
}

// HandlePolicy processes the network policies defined in the SecurityIntent resource.
func (npc *NetworkPolicyController) HandlePolicy(ctx context.Context, intent *intentv1.SecurityIntent) error {
	log := log.FromContext(ctx) // Logger with context.

	// Build and apply/update Cilium Network Policy based on SecurityIntent.
	ciliumPolicy := utils.BuildCiliumNetworkPolicySpec(ctx, intent).(*ciliumv2.CiliumNetworkPolicy)
	err := utils.ApplyOrUpdatePolicy(ctx, npc.Client, ciliumPolicy, ciliumPolicy.Name)
	if err != nil {
		log.Error(err, "Failed to apply Cilium Network Policy", "Name", ciliumPolicy.Name)
		return err
	}

	// If SecurityIntent contains protocol resources, build and apply/update KubeArmor Network Policy.
	if containsProtocolResource(intent) {
		armorNetPolicy := utils.BuildKubeArmorPolicySpec(ctx, intent, utils.GetPolicyType(utils.IsHostPolicy(intent))).(*kubearmorpolicyv1.KubeArmorPolicy)
		err = utils.ApplyOrUpdatePolicy(ctx, npc.Client, armorNetPolicy, armorNetPolicy.Name)
		if err != nil {
			log.Error(err, "Failed to apply KubeArmor Network Policy", "Name", armorNetPolicy.Name)
			return err
		}
	}

	log.Info("Applied Network Policy", "PolicyName", intent.Name)
	return nil
}

// DeletePolicy removes the network policy associated with the SecurityIntent resource.
func (npc *NetworkPolicyController) DeletePolicy(ctx context.Context, intent *intentv1.SecurityIntent) error {
	log := log.FromContext(ctx)
	var err error

	// Delete KubeArmor or Cilium Network Policy based on the contents of SecurityIntent.

	if containsProtocolResource(intent) {
		err = deleteNetworkPolicy(ctx, npc.Client, "KubeArmorPolicy", intent.Name, intent.Namespace)
		if err != nil {
			log.Error(err, "Failed to delete KubeArmor Network Policy", "Name", intent.Name)
			return err
		}
	} else {
		// Delete Cilium Network Policy by default
		err = deleteNetworkPolicy(ctx, npc.Client, "CiliumNetworkPolicy", intent.Name, intent.Namespace)
		if err != nil {
			log.Error(err, "Failed to delete Cilium Network Policy", "Name", intent.Name)
			return err
		}
	}

	log.Info("Deleted Network Policy", "PolicyName", intent.Name)
	return nil
}

// Additional helper functions for policy creation and deletion.
func createCiliumNetworkPolicy(ctx context.Context, intent *intentv1.SecurityIntent) *ciliumv2.CiliumNetworkPolicy {
	return utils.BuildCiliumNetworkPolicySpec(ctx, intent).(*ciliumv2.CiliumNetworkPolicy)
}

func createKubeArmorNetworkPolicy(ctx context.Context, intent *intentv1.SecurityIntent) *kubearmorpolicyv1.KubeArmorPolicy {
	return utils.BuildKubeArmorPolicySpec(ctx, intent, "policy").(*kubearmorpolicyv1.KubeArmorPolicy)
}

func applyCiliumNetworkPolicy(ctx context.Context, c client.Client, policy *ciliumv2.CiliumNetworkPolicy) {
	utils.ApplyOrUpdatePolicy(ctx, c, policy, policy.Name)
}

func applyKubeArmorNetworkPolicy(ctx context.Context, c client.Client, policy *kubearmorpolicyv1.KubeArmorPolicy) {
	utils.ApplyOrUpdatePolicy(ctx, c, policy, policy.Name)
}

// containsProtocolResource checks for the presence of protocol resources in SecurityIntent.
func containsProtocolResource(intent *intentv1.SecurityIntent) bool {
	// Iterates through the intent resources to find if 'protocols' key is present.
	for _, resource := range intent.Spec.Intent.Resource {
		if resource.Key == "protocols" {
			return true
		}
	}
	return false
}

// deleteNetworkPolicy helps in deleting a specified network policy.
func deleteNetworkPolicy(ctx context.Context, c client.Client, policyType, name, namespace string) error {
	// Utilizes utility function to delete the specified network policy.
	return utils.DeletePolicy(ctx, c, policyType, name, namespace)
}
