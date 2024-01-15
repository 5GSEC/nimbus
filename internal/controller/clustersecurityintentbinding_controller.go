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
	"github.com/5GSEC/nimbus/pkg/processor/intentbinder"
	"github.com/5GSEC/nimbus/pkg/processor/policybuilder"
)

// ClusterSecurityIntentBindingReconciler reconciles a ClusterSecurityIntentBinding object
type ClusterSecurityIntentBindingReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=clustersecurityintentbindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=clustersecurityintentbindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=clusternimbuspolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=clusternimbuspolicies/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterSecurityIntentBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	csib := &v1.ClusterSecurityIntentBinding{}
	err := r.Get(ctx, types.NamespacedName{Name: req.Name}, csib)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("ClusterSecurityIntentBinding not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding.Name", csib.Name)
		return ctrl.Result{}, err
	}

	if csib.Status.Status == "" || csib.Status.Status == StatusPending {
		csib.Status.Status = StatusCreated
		if err := r.Status().Update(ctx, csib); err != nil {
			logger.Error(err, "failed to update ClusterSecurityIntentBinding status", "ClusterSecurityIntentBinding.Name", csib.Name)
			return ctrl.Result{}, err
		}
		// Let's re-fetch the ClusterSecurityIntentBinding CR after updating the status
		// so that we have the latest state of the resource on the cluster.
		if err := r.Get(ctx, req.NamespacedName, csib); err != nil {
			logger.Error(err, "failed to re-fetch ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding.Name", csib.Name)
			return ctrl.Result{}, err
		}
		logger.Info("ClusterSecurityIntentBinding found", "ClusterSecurityIntentBinding.Name", csib.Name)
		return ctrl.Result{}, nil
	}

	bindingInfo := intentbinder.MatchAndBindIntentsGlobal(ctx, r.Client, csib)

	clusterNp, err := policybuilder.BuildClusterNimbusPolicy(ctx, r.Client, r.Scheme, bindingInfo)
	if err != nil {
		logger.Error(err, "failed to build ClusterNimbusPolicy")
		return ctrl.Result{}, err
	}

	existingClusterNp := &v1.ClusterNimbusPolicy{}
	err = r.Get(ctx, types.NamespacedName{Name: req.Name}, existingClusterNp)
	if err != nil && errors.IsNotFound(err) {
		if err := r.Create(ctx, clusterNp); err != nil {
			logger.Error(err, "failed to create ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", clusterNp.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	clusterNp.ObjectMeta = existingClusterNp.ObjectMeta
	if err := r.Update(ctx, clusterNp); err != nil {
		logger.Error(err, "failed to update ClusterNimbusPolicy")
		return ctrl.Result{}, err
	}

	if clusterNp.Status.Status == "" || clusterNp.Status.Status == StatusPending {
		clusterNp.Status.Status = StatusCreated
		if err := r.Status().Update(ctx, clusterNp); err != nil {
			logger.Error(err, "failed to update ClusterNimbusPolicy status", "ClusterNimbusPolicy.Name", clusterNp.Name)
			return ctrl.Result{}, err
		}
		logger.Info("ClusterNimbusPolicy created", "ClusterNimbusPolicy.Name", clusterNp.Name)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterSecurityIntentBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.ClusterSecurityIntentBinding{}).
		Owns(&v1.ClusterNimbusPolicy{}).
		Complete(r)
}
