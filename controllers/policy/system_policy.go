package policy

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	utils "github.com/5GSEC/nimbus/controllers/utils"

	intentv1 "github.com/5GSEC/nimbus/api/v1"
	kubearmorhostpolicyv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorHostPolicy/api/security.kubearmor.com/v1"
	kubearmorpolicyv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorPolicy/api/security.kubearmor.com/v1"
)

// SystemPolicyController is a struct to handle system policies.
type SystemPolicyController struct {
	Client client.Client   // Client for interacting with Kubernetes API.
	Scheme *runtime.Scheme // Scheme defines the runtime scheme of the Kubernetes objects.
}

// NewSystemPolicyController creates a new instance of SystemPolicyController.
func NewSystemPolicyController(client client.Client, scheme *runtime.Scheme) *SystemPolicyController {
	return &SystemPolicyController{
		Client: client,
		Scheme: scheme,
	}
}

// HandlePolicy processes the system policy as defined in SecurityIntent.
func (spc *SystemPolicyController) HandlePolicy(ctx context.Context, intent *intentv1.SecurityIntent) error {
	log := log.FromContext(ctx) // Logger with context.

	// Determine if the policy is a HostPolicy.
	isHost := utils.IsHostPolicy(intent)

	var err error
	if isHost {
		// Create and apply a KubeArmorHostPolicy if it's a host policy.
		hostPolicy := createKubeArmorHostPolicy(ctx, intent)
		err = applyKubeArmorHostPolicy(ctx, spc.Client, hostPolicy)
	} else {
		// Create and apply a KubeArmorPolicy otherwise.
		armorPolicy := createKubeArmorPolicy(ctx, intent)
		err = applyKubeArmorPolicy(ctx, spc.Client, armorPolicy)
	}

	if err != nil {
		log.Error(err, "Failed to apply policy")
		return err
	}

	log.Info("Applied System Policy", "PolicyType", utils.GetPolicyType(isHost), "PolicyName", intent.Name)
	return nil
}

// DeletePolicy removes the system policy associated with the SecurityIntent resource.
func (spc *SystemPolicyController) DeletePolicy(ctx context.Context, intent *intentv1.SecurityIntent) error {
	log := log.FromContext(ctx)

	isHost := utils.IsHostPolicy(intent)
	policyType := utils.GetPolicyType(isHost)

	// Delete the system policy.
	err := deleteSystemPolicy(ctx, spc.Client, policyType, intent.Name, intent.Namespace)
	if err != nil {
		log.Error(err, "Failed to delete policy")
		return err
	}

	log.Info("Deleted System Policy", "PolicyType", policyType, "PolicyName", intent.Name)
	return nil
}

// createKubeArmorHostPolicy(): Creates a KubeArmorHostPolicy object based on the given SecurityIntent
func createKubeArmorHostPolicy(ctx context.Context, intent *intentv1.SecurityIntent) *kubearmorhostpolicyv1.KubeArmorHostPolicy {
	return utils.BuildKubeArmorPolicySpec(ctx, intent, "host").(*kubearmorhostpolicyv1.KubeArmorHostPolicy)
}

// createKubeArmorPolicy creates a KubeArmorPolicy object based on the given SecurityIntent
func createKubeArmorPolicy(ctx context.Context, intent *intentv1.SecurityIntent) *kubearmorpolicyv1.KubeArmorPolicy {
	return utils.BuildKubeArmorPolicySpec(ctx, intent, "policy").(*kubearmorpolicyv1.KubeArmorPolicy)
}

// applyKubeArmorPolicy applies a KubeArmorPolicy to the Kubernetes cluster
func applyKubeArmorPolicy(ctx context.Context, c client.Client, policy *kubearmorpolicyv1.KubeArmorPolicy) error {
	return utils.ApplyOrUpdatePolicy(ctx, c, policy, policy.Name)
}

// applyKubeArmorHostPolicy applies a KubeArmorHostPolicy to the Kubernetes cluster
func applyKubeArmorHostPolicy(ctx context.Context, c client.Client, policy *kubearmorhostpolicyv1.KubeArmorHostPolicy) error {
	return utils.ApplyOrUpdatePolicy(ctx, c, policy, policy.Name)
}

func deleteSystemPolicy(ctx context.Context, c client.Client, policyType, name, namespace string) error {
	// Utilizes utility function to delete the specified system policy.
	return utils.DeletePolicy(ctx, c, policyType, name, namespace)
}
