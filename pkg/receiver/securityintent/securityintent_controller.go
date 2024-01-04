// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package securityintent

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/receiver/watcher"
)

type SecurityIntentReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	WatcherController *watcher.WatcherController
}

// NewSecurityIntentReconciler creates a new SecurityIntentReconciler.
func NewSecurityIntentReconciler(client client.Client, scheme *runtime.Scheme) *SecurityIntentReconciler {
	if client == nil {
		fmt.Println("SecurityIntentReconciler: Client is nil")
		return nil
	}

	WatcherController, err := watcher.NewWatcherController(client)
	if err != nil {
		fmt.Println("SecurityIntentReconciler: Failed to initialize WatcherController:", err)
		return nil
	}

	return &SecurityIntentReconciler{
		Client:            client,
		Scheme:            scheme,
		WatcherController: WatcherController,
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

	if r.WatcherController == nil {
		fmt.Println("SecurityIntentReconciler: WatcherController is nil")
		return ctrl.Result{}, fmt.Errorf("WatcherController is not properly initialized")
	}

	intent, err := r.WatcherController.WatcherIntent.Reconcile(ctx, req)
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
