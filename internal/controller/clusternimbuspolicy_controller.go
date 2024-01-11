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

// ClusterNimbusPolicyReconciler reconciles a ClusterNimbusPolicy object
type ClusterNimbusPolicyReconciler struct {
	client.Client
	Scheme                     *runtime.Scheme
	ClusterNimbusPolicyWatcher *watcher.ClusterNimbusPolicy
}

func NewClusterNimbusPolicyReconciler(client client.Client, scheme *runtime.Scheme) *ClusterNimbusPolicyReconciler {
	if client == nil {
		fmt.Println("ClusterNimbusPolicyReconciler.Client is nil")
		return nil
	}

	clusterNPWatcher, err := watcher.NewClusterNimbusPolicy(client)
	if err != nil {
		fmt.Println("failed to initialize ClusterNimbusPolicyWatcher, error:", err)
		return nil
	}
	return &ClusterNimbusPolicyReconciler{
		Client:                     client,
		Scheme:                     scheme,
		ClusterNimbusPolicyWatcher: clusterNPWatcher,
	}
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=clusternimbuspolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=clusternimbuspolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=clusternimbuspolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterNimbusPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	if r.ClusterNimbusPolicyWatcher == nil {
		return ctrl.Result{}, fmt.Errorf("ClusterNimbusPolicyWatcher is not properly initialized")
	}

	cwnp, err := r.ClusterNimbusPolicyWatcher.Reconcile(ctx, req)
	if err != nil {
		logger.Error(err, "failed to reconcile ClusterNimbusPolicy", "ClusterNimbusPolicy", req.Name)
		return ctrl.Result{}, err
	}
	if cwnp != nil {
		logger.Info("ClusterNimbusPolicy found", "Name", req.Name)
	} else {
		logger.Info("ClusterNimbusPolicy not found", "Name", req.Name)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterNimbusPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.ClusterNimbusPolicy{}).
		Complete(r)
}
