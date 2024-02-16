// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controller

import (
	"context"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	v1 "github.com/5GSEC/nimbus/api/v1"
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
	err := r.Get(ctx, req.NamespacedName, csib)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("ClusterSecurityIntentBinding not found. Ignoring since object must be deleted")
			logger.Info("ClusterNimbusPolicy deleted due to ClusterSecurityIntentBinding deletion",
				"ClusterNimbusPolicy.Name", req.Name, "ClusterSecurityIntentBinding.Name", req.Name,
			)
			return doNotRequeue()
		}
		logger.Error(err, "failed to get ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding.Name", csib.Name)
		return requeueWithError(err)
	}
	logger.Info("ClusterSecurityIntentBinding found", "ClusterSecurityIntentBinding.Name", req.Name)

	if err = r.createOrUpdateCwnp(ctx, logger, req); err != nil {
		return requeueWithError(err)
	}

	if err = r.updateStatus(ctx, logger, req); err != nil {
		return requeueWithError(err)
	}

	return doNotRequeue()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterSecurityIntentBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.ClusterSecurityIntentBinding{}).
		Owns(&v1.ClusterNimbusPolicy{}).
		WithEventFilter(
			predicate.Funcs{
				CreateFunc: r.createFn,
				UpdateFunc: r.updateFn,
				DeleteFunc: r.deleteFn,
			},
		).
		Complete(r)
}

func (r *ClusterSecurityIntentBindingReconciler) createFn(createEvent event.CreateEvent) bool {
	if _, ok := createEvent.Object.(*v1.ClusterNimbusPolicy); ok {
		return false
	}
	return true
}

func (r *ClusterSecurityIntentBindingReconciler) updateFn(updateEvent event.UpdateEvent) bool {
	// TODO: Handle update event for ClusterNimbusPolicy update so that reconciler don't process it
	// twice.
	return updateEvent.ObjectOld.GetGeneration() != updateEvent.ObjectNew.GetGeneration()
}

func (r *ClusterSecurityIntentBindingReconciler) deleteFn(deleteEvent event.DeleteEvent) bool {
	obj := deleteEvent.Object
	if _, ok := obj.(*v1.ClusterSecurityIntentBinding); ok {
		return true
	}
	return ownerExists(r.Client, obj)
}

func (r *ClusterSecurityIntentBindingReconciler) createOrUpdateCwnp(ctx context.Context, logger logr.Logger, req ctrl.Request) error {
	// Always fetch the latest CRs so that we have the latest state of the CRs on the
	// cluster.
	var csib v1.ClusterSecurityIntentBinding
	if err := r.Get(ctx, req.NamespacedName, &csib); err != nil {
		logger.Error(err, "failed to fetch ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding.Name", req.Name)
		return err
	}

	var cwnp v1.ClusterNimbusPolicy
	err := r.Get(ctx, req.NamespacedName, &cwnp)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.createCwnp(ctx, logger, csib)
		}
		logger.Error(err, "failed to fetch ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", req.Name)
	}
	return r.updateCwnp(ctx, logger, csib)
}

func (r *ClusterSecurityIntentBindingReconciler) createCwnp(ctx context.Context, logger logr.Logger, csib v1.ClusterSecurityIntentBinding) error {
	clusterNp := policybuilder.BuildClusterNimbusPolicy(ctx, logger, r.Client, r.Scheme, csib)
	if clusterNp == nil {
		return nil
	}

	if err := r.Create(ctx, clusterNp); err != nil {
		logger.Error(err, "failed to create ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", clusterNp.Name)
		return err
	}
	logger.Info("ClusterNimbusPolicy created", "ClusterNimbusPolicy.Name", clusterNp.Name)

	return nil
}

func (r *ClusterSecurityIntentBindingReconciler) updateCwnp(ctx context.Context, logger logr.Logger, csib v1.ClusterSecurityIntentBinding) error {
	var existingCwnp v1.ClusterNimbusPolicy
	if err := r.Get(ctx, types.NamespacedName{Name: csib.Name}, &existingCwnp); err != nil {
		logger.Error(err, "failed to fetch ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", csib.Name)
		return err
	}

	clusterNp := policybuilder.BuildClusterNimbusPolicy(ctx, logger, r.Client, r.Scheme, csib)
	if clusterNp == nil {
		return nil
	}

	clusterNp.ObjectMeta.ResourceVersion = existingCwnp.ObjectMeta.ResourceVersion
	if err := r.Update(ctx, clusterNp); err != nil {
		logger.Error(err, "failed to configure ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", clusterNp.Name)
		return err
	}
	logger.Info("ClusterNimbusPolicy configured", "ClusterNimbusPolicy.Name", clusterNp.Name)

	return nil
}

func (r *ClusterSecurityIntentBindingReconciler) updateStatus(ctx context.Context, logger logr.Logger, req ctrl.Request) error {
	// To handle potential latency issues with the Kubernetes API server, we
	// implement an exponential backoff strategy when fetching the ClusterNimbusPolicy
	// custom resource. This enhances resilience by retrying failed requests with
	// increasing intervals, preventing excessive retries in case of persistent 'Not
	// Found' errors.
	if retryErr := retry.OnError(retry.DefaultRetry, apierrors.IsNotFound, func() error {
		np := &v1.ClusterNimbusPolicy{}
		if err := r.Get(ctx, req.NamespacedName, np); err != nil {
			return err
		}
		return nil
	}); retryErr != nil {
		logger.Error(retryErr, "failed to fetch ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", req.Name)
		return retryErr
	}

	// Since multiple adapters may update the ClusterNimbusPolicy status concurrently,
	// there's a risk of conflict during updates of ClusterNimbusPolicy status. To ensure
	// data consistency, retry on write failures. On conflict, the status update is
	// retried with an exponential backoff strategy. This provides resilience against
	// potential issues while preventing indefinite retries in case of persistent
	// conflicts.
	if retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestCwnp := &v1.ClusterNimbusPolicy{}
		if err := r.Get(ctx, req.NamespacedName, latestCwnp); err != nil {
			return err
		}
		latestCwnp.Status = v1.ClusterNimbusPolicyStatus{
			Status:      StatusCreated,
			LastUpdated: metav1.Now(),
		}
		if err := r.Status().Update(ctx, latestCwnp); err != nil {
			return err
		}
		return nil
	}); retryErr != nil {
		logger.Error(retryErr, "failed to update ClusterNimbusPolicy status", "ClusterNimbusPolicy.Name", req.Name)
		return retryErr
	}

	// Fetch the latest SecurityIntentBinding so that we have the latest state
	// on the cluster.
	latestCsib := &v1.ClusterSecurityIntentBinding{}
	if err := r.Get(ctx, req.NamespacedName, latestCsib); err != nil {
		logger.Error(err, "failed to fetch ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding.Name", req.Name)
		return err
	}
	count, boundIntents := extractBoundIntentsInfo(latestCsib.Spec.Intents)
	latestCsib.Status = v1.ClusterSecurityIntentBindingStatus{
		Status:               StatusCreated,
		LastUpdated:          metav1.Now(),
		NumberOfBoundIntents: count,
		BoundIntents:         boundIntents,
		ClusterNimbusPolicy:  req.Name,
	}
	if err := r.Status().Update(ctx, latestCsib); err != nil {
		logger.Error(err, "failed to update ClusterSecurityIntentBinding status", "ClusterSecurityIntentBinding.Name", latestCsib.Name)
		return err
	}

	return nil
}
