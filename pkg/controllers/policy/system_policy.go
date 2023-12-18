// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package policy

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	general "github.com/5GSEC/nimbus/pkg/controllers/general"
	utils "github.com/5GSEC/nimbus/pkg/controllers/utils"
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
func (spc *SystemPolicyController) HandlePolicy(ctx context.Context, bindingInfo *general.BindingInfo) error {
	log := log.FromContext(ctx) // Logger with context.
	log.Info("Handling System Policy", "BindingName", bindingInfo.Binding.Name)

	// Build KubeArmorPolicy based on BindingInfo
	kubearmorPolicy := utils.BuildKubeArmorPolicySpec(ctx, bindingInfo)

	err := utils.ApplyOrUpdatePolicy(ctx, spc.Client, kubearmorPolicy, bindingInfo.Binding.Name)
	if err != nil {
		log.Error(err, "Failed to apply KubeArmorPolicy", "Name", bindingInfo.Binding.Name)
		return err
	}

	log.Info("Applied KubeArmorPolicy", "PolicyName", bindingInfo.Binding.Name)
	return nil
}

// DeletePolicy removes the system policy associated with the SecurityIntent resource.
func (spc *SystemPolicyController) DeletePolicy(ctx context.Context, bindingInfo *general.BindingInfo) error {
	log := log.FromContext(ctx)

	// Delete KubeArmor Policy
	err := utils.DeletePolicy(ctx, spc.Client, "KubeArmorPolicy", bindingInfo.Binding.Name, bindingInfo.Binding.Namespace)
	if err != nil {
		log.Error(err, "Failed to delete KubeArmor Policy", "Name", bindingInfo.Binding.Name)
		return err
	}

	log.Info("Deleted System Policy", "PolicyName", bindingInfo.Binding.Name)
	return nil
}
