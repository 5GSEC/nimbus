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

type SecurityIntentReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ControllerWatcher *watcher.Controller
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintents/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SecurityIntentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if r.ControllerWatcher == nil {
		return ctrl.Result{}, fmt.Errorf("ControllerWatcher is not properly initialized")
	}

	intent, err := r.ControllerWatcher.SecurityIntentWatcher.Reconcile(ctx, req)
	if err != nil {
		logger.Error(err, "failed to reconcile SecurityIntent", "Request", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if intent != nil {
		logger.Info("SecurityIntent resource found", "Name", req.Name, "Namespace", req.Namespace)
	} else {
		logger.Info("SecurityIntent resource not found", "Name", req.Name, "Namespace", req.Namespace)
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
