// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package nimbuspolicy

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/exporter/httpexporter"
	"github.com/5GSEC/nimbus/pkg/receiver/watcher"
)

// NimbusPolicyReconciler reconciles a NimbusPolicy object.
type NimbusPolicyReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	WatcherNimbusPolicy *watcher.WatcherNimbusPolicy
}

// NewNimbusPolicyReconciler creates a new instance of NimbusPolicyReconciler.
// It initializes the WatcherNimbusPolicy which watches and reacts to changes in NimbusPolicy objects.
func NewNimbusPolicyReconciler(client client.Client, scheme *runtime.Scheme) *NimbusPolicyReconciler {
	if client == nil {
		fmt.Println("NimbusPolicyReconciler: Client is nil")
		return nil
	}

	watcherNimbusPolicy, err := watcher.NewWatcherNimbusPolicy(client)
	if err != nil {
		fmt.Println("NimbusPolicyReconciler: Failed to initialize WatcherNimbusPolicy:", err)
		return nil
	}

	return &NimbusPolicyReconciler{
		Client:              client,
		Scheme:              scheme,
		WatcherNimbusPolicy: watcherNimbusPolicy,
	}
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=nimbuspolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=nimbuspolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=nimbuspolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NimbusPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *NimbusPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if r.WatcherNimbusPolicy == nil {
		fmt.Println("NimbusPolicyReconciler: WatcherNimbusPolicy is nil")
		return ctrl.Result{}, fmt.Errorf("WatcherNimbusPolicy is not properly initialized")
	}

	nimPol, err := r.WatcherNimbusPolicy.Reconcile(ctx, req)
	if err != nil {
		log.Error(err, "Error in WatcherNimbusPolicy.Reconcile", "Request", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if nimPol != nil {
		log.Info("NimbusPolicy resource found", "Name", req.Name, "Namespace", req.Namespace)
	} else {
		log.Info("NimbusPolicy resource not found", "Name", req.Name, "Namespace", req.Namespace)
	}

	// Exporting the NimbusPolicy if it is found.
	if nimPol != nil {
		exporter := httpexporter.NewHttpNimbusExporter("http://localhost:13000/api/v1/nimbus/export") // Update the URL as needed.
		err := exporter.ExportNimbusPolicy(ctx, nimPol)
		if err != nil {
			log.Error(err, "Failed to export NimbusPolicy")
			return ctrl.Result{}, err
		}
		log.Info("NimbusPolicy exported successfully")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
// It registers the NimbusPolicyReconciler to manage NimbusPolicy resources.
func (r *NimbusPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.NimbusPolicy{}).
		Complete(r)
}
