// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package policy

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"

	general "github.com/5GSEC/nimbus/pkg/controllers/general"
	utils "github.com/5GSEC/nimbus/pkg/controllers/utils"
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
func (npc *NetworkPolicyController) HandlePolicy(ctx context.Context, bindingInfo *general.BindingInfo) error {
	log := log.FromContext(ctx)
	log.Info("Handling Network Policy", "BindingName", bindingInfo.Binding.Name)

	// Build and apply/update Cilium Network Policy based on BindingInfo.
	ciliumPolicySpec := utils.BuildCiliumNetworkPolicySpec(ctx, bindingInfo).(*ciliumv2.CiliumNetworkPolicy)
	err := utils.ApplyOrUpdatePolicy(ctx, npc.Client, ciliumPolicySpec, bindingInfo.Binding.Name)
	if err != nil {
		log.Error(err, "Failed to apply Cilium Network Policy", "Name", bindingInfo.Binding.Name)
		return err
	}

	log.Info("Applied Network Policy", "PolicyName", bindingInfo.Binding.Name)
	return nil
}

// DeletePolicy removes the network policy associated with the SecurityIntent resource.
func (npc *NetworkPolicyController) DeletePolicy(ctx context.Context, bindingInfo *general.BindingInfo) error {
	log := log.FromContext(ctx)

	// Modified line: Merged variable declaration with assignment
	err := utils.DeletePolicy(ctx, npc.Client, "CiliumNetworkPolicy", bindingInfo.Binding.Name, bindingInfo.Binding.Namespace)
	if err != nil {
		log.Error(err, "Failed to delete Cilium Network Policy", "Name", bindingInfo.Binding.Name)
		return err
	}

	log.Info("Deleted Network Policy", "PolicyName", bindingInfo.Binding.Name)
	return nil
}
