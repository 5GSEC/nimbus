// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controllers

import (
	"context"
	"fmt"
	
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	intentv1 "github.com/5GSEC/nimbus/Nimbus/api/v1"
	general "github.com/5GSEC/nimbus/Nimbus/controllers/general"
	policy "github.com/5GSEC/nimbus/Nimbus/controllers/policy"
)

// SecurityIntentBindingReconciler reconciles a SecurityIntentBinding object
type SecurityIntentBindingReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	GeneralController *general.GeneralController
	PolicyController  *policy.PolicyController
}

func NewSecurityIntentBindingReconciler(client client.Client, scheme *runtime.Scheme) *SecurityIntentBindingReconciler {
	if client == nil {
		fmt.Println("SecurityIntentBindingReconciler: Client is nil")
		return nil
	}

	generalController, err := general.NewGeneralController(client)
	if err != nil {
		fmt.Println("SecurityIntentBindingReconciler: Failed to initialize GeneralController:", err)
		return nil
	}

	policyController := policy.NewPolicyController(client, scheme)
	if policyController == nil {
		fmt.Println("SecurityIntentBindingReconciler: Failed to initialize PolicyController")
		return nil
	}

	return &SecurityIntentBindingReconciler{
		Client:            client,
		Scheme:            scheme,
		GeneralController: generalController,
		PolicyController:  policyController,
	}
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintentbindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintentbindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintentbindings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecurityIntentBinding object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile

func (r *SecurityIntentBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if r.GeneralController == nil || r.GeneralController.WatcherBinding == nil {
		fmt.Println("SecurityIntentBindingReconciler: GeneralController or WatcherBinding is not initialized")
		return ctrl.Result{}, fmt.Errorf("GeneralController or WatcherBinding is not initialized")
	}

	binding, err := r.GeneralController.WatcherBinding.Reconcile(ctx, req)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("Error in WatcherBinding.Reconcile: %v", err)
	}

	if binding != nil {
		log.Info("SecurityIntentBinding resource found", "Name", req.Name, "Namespace", req.Namespace)

		bindingInfo, err := general.MatchIntentAndBinding(ctx, r.Client, binding)
		if err != nil {
			log.Error(err, "Failed to match SecurityIntent with SecurityIntentBinding", "BindingName", binding.Name)
			return ctrl.Result{}, err
		}

		if bindingInfo != nil {
			err = r.PolicyController.Reconcile(ctx, bindingInfo)
			if err != nil {
				log.Error(err, "Failed to apply policy for SecurityIntentBinding", "BindingName", binding.Name)
				return ctrl.Result{}, err
			}
		} else {
			log.Info("No matching SecurityIntent found for SecurityIntentBinding", "BindingName", binding.Name)
		}
	} else {
		log.Info("SecurityIntentBinding resource not found", "Name", req.Name, "Namespace", req.Namespace)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecurityIntentBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&intentv1.SecurityIntentBinding{}).
		Complete(r)
}
