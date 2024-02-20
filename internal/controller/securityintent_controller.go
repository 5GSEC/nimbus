// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

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
		if apierrors.IsNotFound(err) {
			logger.Info("SecurityIntent not found. Ignoring since object must be deleted")
			return doNotRequeue()
		}
		logger.Error(err, "failed to fetch SecurityIntent", "SecurityIntent.Name", req.Name)
		return requeueWithError(err)
	}

	if err = r.updateStatus(ctx, req.Name); err != nil {
		logger.Error(err, "failed to update SecurityIntent status", "SecurityIntent.Name", req.Name)
		return requeueWithError(err)
	}
	logger.Info("SecurityIntent found", "SecurityIntent.Name", si.Name)

	return doNotRequeue()
}

// SetupWithManager sets up the reconciler with the provided manager.
func (r *SecurityIntentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.SecurityIntent{}).
		WithEventFilter(
			predicate.GenerationChangedPredicate{},
		).
		Complete(r)
}

func (r *SecurityIntentReconciler) updateStatus(ctx context.Context, name string) error {
	latestSi := &v1.SecurityIntent{}
	if getErr := r.Get(ctx, types.NamespacedName{Name: name}, latestSi); getErr != nil {
		return getErr
	}
	latestSi.Status = v1.SecurityIntentStatus{
		ID:     latestSi.Spec.Intent.ID,
		Action: latestSi.Spec.Intent.Action,
		Status: StatusCreated,
	}
	return r.Status().Update(ctx, latestSi)
}
