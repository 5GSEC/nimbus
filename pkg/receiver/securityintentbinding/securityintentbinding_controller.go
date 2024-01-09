// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package securityintentbinding

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
	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
)

// SecurityIntentBindingReconciler reconciles a SecurityIntentBinding object
type SecurityIntentBindingReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	WatcherController *watcher.WatcherController
}

func NewSecurityIntentBindingReconciler(client client.Client, scheme *runtime.Scheme) *SecurityIntentBindingReconciler {
	if client == nil {
		fmt.Println("SecurityIntentBindingReconciler: Client is nil")
		return nil
	}

	WatcherController, err := watcher.NewWatcherController(client)
	if err != nil {
		fmt.Println("SecurityIntentBindingReconciler: Failed to initialize WatcherController:", err)
		return nil
	}

	return &SecurityIntentBindingReconciler{
		Client:            client,
		Scheme:            scheme,
		WatcherController: WatcherController,
	}
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintentbindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintentbindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=securityintentbindings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecurityIntentBinding object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile

func (r *SecurityIntentBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if r.WatcherController == nil {
		fmt.Println("SecurityIntentBindingReconciler: WatcherController is nil")
		return ctrl.Result{}, fmt.Errorf("WatcherController is not properly initialized")
	}

	binding, err := r.WatcherController.WatcherBinding.Reconcile(ctx, req)
	if err != nil {
		log.Error(err, "Error in WatcherBinding.Reconcile", "Request", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if binding != nil {
		log.Info("SecurityIntentBinding resource found", "Name", req.Name, "Namespace", req.Namespace)
	} else {
		log.Info("SecurityIntentBinding resource not found", "Name", req.Name, "Namespace", req.Namespace)

		// Delete associated NimbusPolicy if exists
		nimbusPolicy := &v1.NimbusPolicy{}
		err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, nimbusPolicy)
		if err != nil && !errors.IsNotFound(err) {
			log.Error(err, "Failed to get NimbusPolicy for deletion")
			return ctrl.Result{}, err
		}
		if err == nil {
			// NimbusPolicy exists, delete it
			if err := r.Delete(ctx, nimbusPolicy); err != nil {
				log.Error(err, "Failed to delete NimbusPolicy")
				return ctrl.Result{}, err
			}
			log.Info("Deleted NimbusPolicy due to SecurityIntentBinding deletion", "NimbusPolicy", req.NamespacedName)
		}
		// Delete Kubearmor Policy with the same name and namespace
		kubearmorPolicy := &kubearmorv1.KubeArmorPolicy{}
		if err := r.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, kubearmorPolicy); err == nil {
			if err := r.Delete(ctx, kubearmorPolicy); err != nil {
				log.Error(err, "Failed to delete KubearmorPolicy")
				return ctrl.Result{}, err
			}
			log.Info("Deleted KubearmorPolicy due to SecurityIntentBinding deletion", "KubearmorPolicy", req.NamespacedName)
		}
		return ctrl.Result{}, nil
	}

	// Call the MatchAndBindIntents function to generate the binding information.
	bindingInfo, err := intentbinder.MatchAndBindIntents(ctx, r.Client, req, binding)
	if err != nil {
		log.Error(err, "Failed to match and bind intents")
		return ctrl.Result{}, err
	}

	// Create a NimbusPolicy.
	nimbusPolicy, err := nimbuspolicybuilder.BuildNimbusPolicy(ctx, r.Client, req, bindingInfo)
	if err != nil {
		log.Error(err, "Failed to build NimbusPolicy")
		return ctrl.Result{}, err
	}

	// Store the NimbusPolicy on the Kubernetes API server.
	if err := r.Create(ctx, nimbusPolicy); err != nil {
		log.Error(err, "Failed to create NimbusPolicy")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecurityIntentBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.SecurityIntentBinding{}).
		Complete(r)
}
