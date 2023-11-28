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

package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/5GSEC/nimbus/api/v1"
	general "github.com/5GSEC/nimbus/controllers/general"
	policy "github.com/5GSEC/nimbus/controllers/policy"
)

// SecurityIntentReconciler reconciles a SecurityIntent object.
type SecurityIntentReconciler struct {
	client.Client
	Scheme            *runtime.Scheme            // Scheme defines the runtime scheme of the Kubernetes objects.
	GeneralController *general.GeneralController // GeneralController is a custom controller for general operations.
	PolicyController  *policy.PolicyController   // PolicyController is a custom controller for policy operations.
}

// NewSecurityIntentReconciler creates a new SecurityIntentReconciler.
func NewSecurityIntentReconciler(client client.Client, scheme *runtime.Scheme) *SecurityIntentReconciler {
	// Check if the client is nil.
	if client == nil {
		fmt.Println("SecurityIntentReconciler: Client is nil")
		return nil
	}

	// Initialize GeneralController; if failed, return nil.
	generalController, err := general.NewGeneralController(client)
	if err != nil || generalController == nil { // Check if generalController is nil.
		fmt.Println("SecurityIntentReconciler: Failed to initialize GeneralController:", err)
		return nil
	}

	// Initialize PolicyController; if failed, return nil.
	policyController := policy.NewPolicyController(client, scheme)
	if policyController == nil {
		fmt.Println("SecurityIntentReconciler: Failed to initialize PolicyController")
		return nil
	}

	// Return a new instance of SecurityIntentReconciler.
	return &SecurityIntentReconciler{
		Client:            client,
		Scheme:            scheme,
		GeneralController: generalController,
		PolicyController:  policyController,
	}
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

// Reconcile handles the reconciliation of the SecurityIntent resources.
func (r *SecurityIntentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Check if GeneralController or its components are properly initialized.
	if r.GeneralController == nil {
		fmt.Println("SecurityIntentReconciler: GeneralController is nil")
		return ctrl.Result{}, fmt.Errorf("GeneralController is not properly initialized")
	}
	if r.GeneralController.WatcherIntent == nil {
		fmt.Println("SecurityIntentReconciler: WatcherIntent is nil")
		return ctrl.Result{}, fmt.Errorf("WatcherIntent is not properly initialized")
	}

	// Perform Reconcile logic regardless of the state of GeneralController and WatcherIntent.
	intent, err := r.GeneralController.Reconcile(ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}

	if intent == nil {
		return ctrl.Result{}, nil
	}

	// Invoke the PolicyController's Reconcile method with the intent.
	err = r.PolicyController.Reconcile(ctx, intent)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the reconciler with the provided manager.
func (r *SecurityIntentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Set up the controller to manage SecurityIntent resources.
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.SecurityIntent{}).
		Complete(r)
}
