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

	"github.com/5GSEC/nimbus/pkg/watcher"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/processor/intentbinder"
	"github.com/5GSEC/nimbus/pkg/processor/policybuilder"
)

// SecurityIntentBindingReconciler reconciles a SecurityIntentBinding object
type SecurityIntentBindingReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ControllerWatcher *watcher.Controller
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintentbindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintentbindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintentbindings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SecurityIntentBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if r.ControllerWatcher == nil {
		return ctrl.Result{}, fmt.Errorf("ControllerWatcher is not properly initialized")
	}

	binding, err := r.ControllerWatcher.SecurityIntentBindingWatcher.Reconcile(ctx, req)
	if err != nil {
		logger.Error(err, "failed to reconcile SecurityIntentBinding", "Request", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if binding != nil {
		logger.Info("SecurityIntentBinding resource found", "Name", req.Name, "Namespace", req.Namespace)
	} else {
		logger.Info("SecurityIntentBinding resource not found", "Name", req.Name, "Namespace", req.Namespace)

		// Delete associated NimbusPolicy if exists
		nimbusPolicy := &v1.NimbusPolicy{}
		err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, nimbusPolicy)
		if err != nil && !errors.IsNotFound(err) {
			logger.Error(err, "failed to get NimbusPolicy for deletion")
			return ctrl.Result{}, err
		}
		if err == nil {
			// NimbusPolicy exists, delete it
			if err := r.Delete(ctx, nimbusPolicy); err != nil {
				logger.Error(err, "failed to delete NimbusPolicy")
				return ctrl.Result{}, err
			}
			logger.Info("Deleted NimbusPolicy due to SecurityIntentBinding deletion", "NimbusPolicy", req.NamespacedName)
		}
		//Todo: Signal adapters to delete corresponding policies.
		return ctrl.Result{}, nil
	}

	// Call the MatchAndBindIntents function to generate the binding information.
	bindingInfo, err := intentbinder.MatchAndBindIntents(ctx, r.Client, binding)
	if err != nil {
		logger.Error(err, "failed to match and bind intents")
		return ctrl.Result{}, err
	}

	// Create a NimbusPolicy.
	nimbusPolicy, err := policybuilder.BuildNimbusPolicy(ctx, r.Client, bindingInfo)
	if err != nil {
		logger.Error(err, "failed to build NimbusPolicy")
		return ctrl.Result{}, err
	}

	// Store the NimbusPolicy on the Kubernetes API server.
	if err := r.Create(ctx, nimbusPolicy); err != nil {
		logger.Error(err, "Failed to create NimbusPolicy")
		return ctrl.Result{}, err
	}
	//Todo: Update status
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecurityIntentBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.SecurityIntentBinding{}).
		Complete(r)
}
