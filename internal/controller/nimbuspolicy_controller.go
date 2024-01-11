// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/watcher"
)

// NimbusPolicyReconciler reconciles a NimbusPolicy object.
type NimbusPolicyReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	NimbusPolicyWatcher *watcher.NimbusPolicy
}

// NewNimbusPolicyReconciler creates a new instance of NimbusPolicyReconciler.
// It initializes the WatcherNimbusPolicy which watches and reacts to changes in NimbusPolicy objects.
func NewNimbusPolicyReconciler(client client.Client, scheme *runtime.Scheme) *NimbusPolicyReconciler {
	if client == nil {
		fmt.Println("NimbusPolicyReconciler.Client is nil")
		return nil
	}

	nimbusPolicyWatcher, err := watcher.NewNimbusPolicy(client)
	if err != nil {
		fmt.Println("failed to initialize NimbusPolicyWatcher, error:", err)
		return nil
	}

	return &NimbusPolicyReconciler{
		Client:              client,
		Scheme:              scheme,
		NimbusPolicyWatcher: nimbusPolicyWatcher,
	}
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=nimbuspolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=nimbuspolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=nimbuspolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NimbusPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if r.NimbusPolicyWatcher == nil {
		return ctrl.Result{}, fmt.Errorf("NimbusPolicyWatcher is not properly initialized")
	}

	nimPol, err := r.NimbusPolicyWatcher.Reconcile(ctx, req)
	if err != nil {
		logger.Error(err, "failed to reconcile NimbusPolicy", "NimbusPolicy", req.Name)
		return ctrl.Result{}, err
	}

	if nimPol != nil {
		logger.Info("NimbusPolicy found", "Name", req.Name, "Namespace", req.Namespace)
	} else {
		logger.Info("NimbusPolicy not found", "Name", req.Name, "Namespace", req.Namespace)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
// It registers the NimbusPolicyReconciler to manage NimbusPolicy resources.
func (r *NimbusPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.NimbusPolicy{}).
		Complete(r)
}
