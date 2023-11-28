package policy

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	intentv1 "github.com/5GSEC/nimbus/api/v1"
	utils "github.com/5GSEC/nimbus/controllers/utils"
)

// Constant for the finalizer name used in the SecurityIntent resource.
const securityIntentFinalizer = "finalizer.securityintent.intent.security.nimbus.com"

// PolicyController struct handles different types of policies.
type PolicyController struct {
	Client                  client.Client            // Client for interacting with Kubernetes API.
	Scheme                  *runtime.Scheme          // Scheme defines the runtime scheme of the Kubernetes objects.
	SystemPolicyController  *SystemPolicyController  // Controller for handling system policies.
	NetworkPolicyController *NetworkPolicyController // Controller for handling network policies.
}

// NewPolicyController creates a new instance of PolicyController.
func NewPolicyController(client client.Client, scheme *runtime.Scheme) *PolicyController {
	if client == nil || scheme == nil {
		// Print an error and return nil if the client or scheme is not provided.
		fmt.Println("PolicyController: Client or Scheme is nil")
		return nil
	}

	// Initialize and return a new PolicyController with system and network policy controllers.
	return &PolicyController{
		Client:                  client,
		Scheme:                  scheme,
		SystemPolicyController:  NewSystemPolicyController(client, scheme),
		NetworkPolicyController: NewNetworkPolicyController(client, scheme),
	}
}

// Reconcile handles the reconciliation logic for the SecurityIntent resource.
func (pc *PolicyController) Reconcile(ctx context.Context, intent *intentv1.SecurityIntent) error {
	log := log.FromContext(ctx) // Logger with context.
	log.Info("Processing policy", "Name", intent.Name, "Type", intent.Spec.Intent.Type)

	var err error

	// Switch-case to handle different types of policies based on the intent type.
	switch intent.Spec.Intent.Type {
	case "system":
		log.Info("Handling system policy")
		err = pc.SystemPolicyController.HandlePolicy(ctx, intent)
	case "network":
		log.Info("Handling network policy")
		err = pc.NetworkPolicyController.HandlePolicy(ctx, intent)
	default:
		log.Info("Unknown policy type", "Type", intent.Spec.Intent.Type)
	}

	// Handling finalizer logic for clean up during delete operations.
	if intent.ObjectMeta.DeletionTimestamp.IsZero() {
		// If the resource is not being deleted, add the finalizer if it's not present.
		if !utils.ContainsString(intent.ObjectMeta.Finalizers, securityIntentFinalizer) {
			intent.ObjectMeta.Finalizers = append(intent.ObjectMeta.Finalizers, securityIntentFinalizer)
			err = pc.Client.Update(ctx, intent)
		}
	} else {
		// If the resource is being deleted, process deletion based on policy type and remove finalizer.
		if utils.ContainsString(intent.ObjectMeta.Finalizers, securityIntentFinalizer) {
			switch intent.Spec.Intent.Type {
			case "system":
				err = pc.SystemPolicyController.DeletePolicy(ctx, intent)
			case "network":
				err = pc.NetworkPolicyController.DeletePolicy(ctx, intent)
			default:
				err = fmt.Errorf("unknown policy type: %s", intent.Spec.Intent.Type)
			}

			// Removing the finalizer after handling deletion.
			intent.ObjectMeta.Finalizers = utils.RemoveString(intent.ObjectMeta.Finalizers, securityIntentFinalizer)
			if updateErr := pc.Client.Update(ctx, intent); updateErr != nil {
				return updateErr
			}
		}
	}

	return err
}
