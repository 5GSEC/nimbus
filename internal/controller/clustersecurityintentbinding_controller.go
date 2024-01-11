// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/processor/intentbinder"
	"github.com/5GSEC/nimbus/pkg/processor/nimbuspolicybuilder"
	"github.com/5GSEC/nimbus/pkg/receiver/watcher"
)

// ClusterSecurityIntentBindingReconciler reconciles a ClusterSecurityIntentBinding object
type ClusterSecurityIntentBindingReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	WatcherController *watcher.WatcherController
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=clustersecurityintentbindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=clustersecurityintentbindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=clustersecurityintentbindings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterSecurityIntentBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	if r.WatcherController == nil {
		logger.Info("ClusterSecurityIntentBindingReconciler.WatcherController is nil", "WatcherController", r.WatcherController)
		return ctrl.Result{}, fmt.Errorf("WatcherController is not properly initialized")
	}

	// Todo: Change "Watcher" prefix to suffix.
	clusterBinding, err := r.WatcherController.WatcherClusterBinding.Reconcile(ctx, req)
	if err != nil {
		logger.Error(err, "failed to reconcile ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding", req.Name)
		return ctrl.Result{}, err
	}

	if clusterBinding != nil {
		logger.Info("ClusterSecurityIntentBinding found", "Name", req.Name)
	} else {
		logger.Info("ClusterSecurityIntentBinding not found", "Name", req.Name)
		// Delete associated ClusterNimbusPolicy if exists.
		var clusterNp v1.ClusterNimbusPolicy
		err = r.Get(ctx, types.NamespacedName{Name: req.Name}, &clusterNp)
		if errors.IsNotFound(err) {
			logger.Error(err, "failed to get ClusterNimbusPolicy for deletion", "ClusterNimbusPolicy", clusterNp.Name)
			return ctrl.Result{}, err
		}
		if err == nil {
			if err = r.Delete(ctx, &clusterNp); err != nil {
				logger.Error(err, "failed to delete ClusterNimbusPolicy for deletion", "ClusterNimbusPolicy", clusterNp.Name)
				return ctrl.Result{}, err
			}
		}
		logger.Info("Deleted ClusterNimbusPolicy due to ClusterSecurityIntentBinding deletion", "ClusterNimbusPolicy", clusterNp.Name)
		//Todo: Signal adapters to delete corresponding policies.
		return ctrl.Result{}, nil
	}

	clusterBindingInfo, err := intentbinder.MatchAndBindIntentsGlobal(ctx, r.Client, clusterBinding)
	if err != nil {
		logger.Error(err, "failed to match and bind intents")
		return ctrl.Result{}, err
	}

	cwnp, err := nimbuspolicybuilder.BuildClusterNimbusPolicy(ctx, r.Client, clusterBindingInfo)
	if err != nil {
		logger.Error(err, "failed to build ClusterNimbusPolicy")
		return ctrl.Result{}, err
	}

	if err = r.Create(ctx, cwnp); err != nil {
		logger.Error(err, "failed to create ClusterNimbusPolicy", "ClusterNimbusPolicy", cwnp.Name)
		return ctrl.Result{}, err
	}
	// Todo: Update Status
	//cwnp.Status = v1.ClusterNimbusPolicyStatus{Status: "Created"}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterSecurityIntentBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.ClusterSecurityIntentBinding{}).
		Complete(r)
}
