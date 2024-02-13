// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/processor/intentbinder"
	"github.com/5GSEC/nimbus/pkg/processor/policybuilder"
)

// SecurityIntentBindingReconciler reconciles a SecurityIntentBinding object
type SecurityIntentBindingReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintentbindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintentbindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=nimbuspolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=nimbuspolicies/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SecurityIntentBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	sib := &v1.SecurityIntentBinding{}
	err := r.Get(ctx, req.NamespacedName, sib)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("SecurityIntentBinding not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get SecurityIntentBinding", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
		return ctrl.Result{}, err
	}

	if sib.Status.Status == "" {
		sib.Status.Status = StatusCreated
		sib.Status.LastUpdated = metav1.Now()
		if err := r.Status().Update(ctx, sib); err != nil {
			logger.Error(err, "failed to update SecurityIntentBinding status", "SecurityIntentBinding.Name", sib.Name, "SecurityIntentBinding.Namespace", sib.Namespace)
			return ctrl.Result{}, err
		}
		// Let's re-fetch the SecurityIntentBinding CR after updating the status so that
		// we have the latest state of the resource on the cluster.
		if err := r.Get(ctx, req.NamespacedName, sib); err != nil {
			logger.Error(err, "failed to re-fetch SecurityIntentBinding", "SecurityIntentBinding.Name", sib.Name, "SecurityIntentBinding.Namespace", sib.Namespace)
			return ctrl.Result{}, err
		}
		logger.Info("SecurityIntentBinding found", "SecurityIntentBinding.Name", sib.Name, "SecurityIntentBinding.Namespace", sib.Namespace)
		return ctrl.Result{}, nil
	}

	bindingInfo := intentbinder.MatchAndBindIntents(ctx, r.Client, sib)
	nimbusPolicy, err := policybuilder.BuildNimbusPolicy(ctx, r.Client, r.Scheme, bindingInfo)
	if err != nil {
		logger.Error(err, "failed to build NimbusPolicy")
		return ctrl.Result{}, err
	}

	existingNp := &v1.NimbusPolicy{}
	err = r.Get(ctx, req.NamespacedName, existingNp)
	if err != nil && errors.IsNotFound(err) {
		if err := r.Create(ctx, nimbusPolicy); err != nil {
			logger.Error(err, "failed to create NimbusPolicy", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)
			return ctrl.Result{}, err
		}
		logger.Info("NimbusPolicy created", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)
	} else if err == nil {
		nimbusPolicy.ObjectMeta.ResourceVersion = existingNp.ObjectMeta.ResourceVersion

		// Check if np needs to be updated
		if sibChanged(sib, existingNp) {
			nimbusPolicy.Status.LastUpdated = metav1.Now()
			if err := r.Update(ctx, nimbusPolicy); err != nil {
				logger.Error(err, "failed to update NimbusPolicy")
				return ctrl.Result{}, err
			}
			logger.Info("NimbusPolicy updated", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)

			if err := r.updateSIBStatus(ctx, sib); err != nil {
				logger.Error(err, "failed to update SecurityIntentBinding status after NimbusPolicy operation", "SecurityIntentBinding.Name", sib.Name, "SecurityIntentBinding.Namespace", sib.Namespace)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
	}

	if nimbusPolicy.Status.Status == "" || nimbusPolicy.Status.Status == StatusPending {
		nimbusPolicy.Status.Status = StatusCreated
		nimbusPolicy.Status.LastUpdated = metav1.Now()
		if err := r.Status().Update(ctx, nimbusPolicy); err != nil {
			logger.Error(err, "failed to update NimbusPolicy status", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)
			return ctrl.Result{}, err
		}
		logger.Info("NimbusPolicy created", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)
	}
	return ctrl.Result{}, nil
}

func sibChanged(sib *v1.SecurityIntentBinding, np *v1.NimbusPolicy) bool {
	return sib.Status.LastUpdated.Time.Before(np.Status.LastUpdated.Time)
}

func (r *SecurityIntentBindingReconciler) updateSIBStatus(ctx context.Context, sib *v1.SecurityIntentBinding) error {
	sib.Status.LastUpdated = metav1.Now()
	return r.Status().Update(ctx, sib)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecurityIntentBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.SecurityIntentBinding{}).
		Owns(&v1.NimbusPolicy{}).
		Complete(r)
}
