// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
)

type SecurityIntentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintents/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SecurityIntentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	si := &v1.SecurityIntent{}

	err := r.Get(ctx, types.NamespacedName{Name: req.Name}, si)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("SecurityIntent not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get SecurityIntent", "SecurityIntent.Name", req.Name)
		return ctrl.Result{}, err
	}

	if si.Status.Status == "" || si.Status.Status == StatusPending {
		si.Status.Status = StatusCreated
		if err = r.Status().Update(ctx, si); err != nil {
			logger.Error(err, "failed to update SecurityIntent status", "SecurityIntent.Name", si.Name)
			return ctrl.Result{}, err
		}

		// Let's re-fetch the SecurityIntent Custom Resource after updating the status so
		// that we have the latest state of the resource on the cluster.
		if err = r.Get(ctx, types.NamespacedName{Name: si.Name, Namespace: si.Namespace}, si); err != nil {
			logger.Error(err, "failed to re-fetch SecurityIntent", "SecurityIntent.Name", si.Name)
			return ctrl.Result{}, err
		}

		logger.Info("SecurityIntent found", "SecurityIntent.Name", si.Name)
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
