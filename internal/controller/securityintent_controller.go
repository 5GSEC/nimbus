// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			// When SI is deleted, we should trigger update for related SIBs
			if err := r.updateRelatedSIBs(ctx, req); err != nil {
				logger.Error(err, "failed to update related SecurityIntentBindings after SI deletion", "SecurityIntent.Name", req.Name)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get SecurityIntent", "SecurityIntent.Name", req.Name)
		return ctrl.Result{}, err
	}

	var sibList v1.SecurityIntentBindingList
	if err := r.List(ctx, &sibList, client.InNamespace(req.Namespace)); err != nil {
		logger.Error(err, "unable to list SecurityIntentBindings for update")
		return ctrl.Result{}, err
	}

	for i := range sibList.Items {
		sib := &sibList.Items[i]
		for _, intentRef := range sib.Spec.Intents {
			if intentRef.Name == req.Name {
				sib.Status.LastUpdated = metav1.Now()
				if err := r.Status().Update(ctx, sib); err != nil {
					logger.Error(err, "failed to update SecurityIntentBinding status for SI update", "SecurityIntentBinding.Name", sib.Name)
					return ctrl.Result{}, err
				}
				logger.Info("Updated SecurityIntentBinding due to SecurityIntent change", "SecurityIntentBinding", sib.Name, "SecurityIntent", req.Name)
				break
			}
		}
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

// Update related SecurityIntentBindings after SecurityIntent deletion
func (r *SecurityIntentReconciler) updateRelatedSIBs(ctx context.Context, req ctrl.Request) error {
	var sibList v1.SecurityIntentBindingList
	if err := r.List(ctx, &sibList, client.InNamespace(req.Namespace)); err != nil {
		return err
	}

	logger := log.FromContext(ctx)

	for _, sib := range sibList.Items {
		sibCopy := sib
		updated := false
		for idx, intentRef := range sibCopy.Spec.Intents {
			if intentRef.Name == req.Name {
				// Remove the reference to the deleted or updated SecurityIntent
				sibCopy.Spec.Intents = append(sibCopy.Spec.Intents[:idx], sibCopy.Spec.Intents[idx+1:]...)
				updated = true
				break
			}
		}
		if updated {
			// Mark SIB as needing an update
			if err := r.Update(ctx, &sibCopy); err != nil { // 수정된 복사본 사용
				logger.Error(err, "Failed to update SecurityIntentBinding after SI deletion/update", "SecurityIntentBinding.Name", sibCopy.Name)
				return err
			}
			logger.Info("Updated SecurityIntentBinding due to SecurityIntent deletion/update", "SecurityIntentBinding", sibCopy.Name, "SecurityIntent", req.Name)
		}
	}

	return nil
}

// SetupWithManager sets up the reconciler with the provided manager.
func (r *SecurityIntentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Set up the controller to manage SecurityIntent resources.
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.SecurityIntent{}).
		Complete(r)
}
