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

	"github.com/5GSEC/nimbus/pkg/api/v1"
	general "github.com/5GSEC/nimbus/pkg/controllers/general"
)

type SecurityIntentReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	GeneralController *general.GeneralController
}

// NewSecurityIntentReconciler creates a new SecurityIntentReconciler.
func NewSecurityIntentReconciler(client client.Client, scheme *runtime.Scheme) *SecurityIntentReconciler {
	if client == nil {
		fmt.Println("SecurityIntentReconciler: Client is nil")
		return nil
	}

	generalController, err := general.NewGeneralController(client)
	if err != nil {
		fmt.Println("SecurityIntentReconciler: Failed to initialize GeneralController:", err)
		return nil
	}

	return &SecurityIntentReconciler{
		Client:            client,
		Scheme:            scheme,
		GeneralController: generalController,
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
	log := log.FromContext(ctx)

	if r.GeneralController == nil {
		fmt.Println("SecurityIntentReconciler: GeneralController is nil")
		return ctrl.Result{}, fmt.Errorf("GeneralController is not properly initialized")
	}

	intent, err := r.GeneralController.WatcherIntent.Reconcile(ctx, req)
	if err != nil {
		log.Error(err, "Error in WatcherIntent.Reconcile", "Request", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if intent != nil {
		log.Info("SecurityIntent resource found", "Name", req.Name, "Namespace", req.Namespace)
	} else {
		log.Info("SecurityIntent resource not found", "Name", req.Name, "Namespace", req.Namespace)
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
