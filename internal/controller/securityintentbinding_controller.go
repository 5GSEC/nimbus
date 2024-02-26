// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controller

import (
	"context"
	"strings"

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
		if apierrors.IsNotFound(err) {
			logger.Info("SecurityIntentBinding not found. Ignoring since object must be deleted")
			logger.Info("NimbusPolicy deleted due to SecurityIntentBinding deletion",
				"NimbusPolicy.Name", req.Name, "NimbusPolicy.Namespace", req.Namespace,
				"SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
			return doNotRequeue()
		}
		logger.Error(err, "failed to fetch SecurityIntentBinding", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
		return requeueWithError(err)
	}
	logger.Info("SecurityIntentBinding found", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)

	if np, err := r.createOrUpdateNp(ctx, logger, req); err != nil {
		return requeueWithError(err)
	} else if np == nil {
		return doNotRequeue()
	}

	if err = r.updateStatus(ctx, logger, req); err != nil {
		return requeueWithError(err)
	}

	return doNotRequeue()
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecurityIntentBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.SecurityIntentBinding{}).
		Owns(&v1.NimbusPolicy{}).
		WithEventFilter(
			predicate.Funcs{
				CreateFunc: r.createFn,
				UpdateFunc: r.updateFn,
				DeleteFunc: r.deleteFn,
			},
		).
		Complete(r)
}

func (r *SecurityIntentBindingReconciler) createFn(createEvent event.CreateEvent) bool {
	if _, ok := createEvent.Object.(*v1.NimbusPolicy); ok {
		return false
	}
	return true
}

func (r *SecurityIntentBindingReconciler) updateFn(updateEvent event.UpdateEvent) bool {
	// TODO: Handle update event for NimbusPolicy update so that reconciler don't process it
	// twice.
	return updateEvent.ObjectOld.GetGeneration() != updateEvent.ObjectNew.GetGeneration()
}

func (r *SecurityIntentBindingReconciler) deleteFn(deleteEvent event.DeleteEvent) bool {
	obj := deleteEvent.Object
	if _, ok := obj.(*v1.SecurityIntentBinding); ok {
		return true
	}
	return ownerExists(r.Client, obj)
}

func (r *SecurityIntentBindingReconciler) createOrUpdateNp(ctx context.Context, logger logr.Logger, req ctrl.Request) (*v1.NimbusPolicy, error) { // Always fetch the latest CRs so that we have the latest state of the CRs on the
	// cluster.

	var sib v1.SecurityIntentBinding
	if err := r.Get(ctx, req.NamespacedName, &sib); err != nil {
		logger.Error(err, "failed to fetch SecurityIntentBinding", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
		return nil, err
	}

	var np v1.NimbusPolicy
	err := r.Get(ctx, req.NamespacedName, &np)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.createNp(ctx, logger, sib)
		}
		logger.Error(err, "failed to fetch NimbusPolicy", "NimbusPolicy.Name", req.Name, "NimbusPolicy.Namespace", req.Namespace)
		return nil, err
	}
	return r.updateNp(ctx, logger, sib)
}

func (r *SecurityIntentBindingReconciler) createNp(ctx context.Context, logger logr.Logger, sib v1.SecurityIntentBinding) (*v1.NimbusPolicy, error) {
	nimbusPolicy, err := policybuilder.BuildNimbusPolicy(ctx, logger, r.Client, r.Scheme, sib)
	// TODO: Improve error handling for CEL
	if err != nil {
		// If error is caused due to CEL then we don't retry to build NimbusPolicy.
		if strings.Contains(err.Error(), "error processing CEL") {
			logger.Error(err, "failed to build NimbusPolicy")
			return nil, nil
		}
		logger.Error(err, "failed to build NimbusPolicy")
		return nil, err
	}
	if nimbusPolicy == nil {
		logger.Info("Abort NimbusPolicy creation as no labels matched the CEL expressions")
		return nil, nil
	}

	if err := r.Create(ctx, nimbusPolicy); err != nil {
		logger.Error(err, "failed to create NimbusPolicy", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)
		return nil, err
	}
	logger.Info("NimbusPolicy created", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)

	return nimbusPolicy, nil
}

func (r *SecurityIntentBindingReconciler) updateNp(ctx context.Context, logger logr.Logger, sib v1.SecurityIntentBinding) (*v1.NimbusPolicy, error) {
	var existingNp v1.NimbusPolicy
	if err := r.Get(ctx, types.NamespacedName{Name: sib.Name, Namespace: sib.Namespace}, &existingNp); err != nil {
		logger.Error(err, "failed to fetch NimbusPolicy", "NimbusPolicy.Name", sib.Name, "NimbusPolicy.Namespace", sib.Namespace)
		return nil, err
	}

	nimbusPolicy, err := policybuilder.BuildNimbusPolicy(ctx, logger, r.Client, r.Scheme, sib)
	// TODO: Improve error handling for CEL
	if err != nil {
		// If error is caused due to CEL then we don't retry to build NimbusPolicy.
		if strings.Contains(err.Error(), "error processing CEL") {
			logger.Error(err, "failed to build NimbusPolicy")
			return nil, nil
		}
		logger.Error(err, "failed to build NimbusPolicy")
		return nil, err
	}
	if nimbusPolicy == nil {
		logger.Info("Abort NimbusPolicy creation as no labels matched the CEL expressions")
		return nil, nil
	}

	nimbusPolicy.ObjectMeta.ResourceVersion = existingNp.ObjectMeta.ResourceVersion
	if err := r.Update(ctx, nimbusPolicy); err != nil {
		logger.Error(err, "failed to configure NimbusPolicy", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)
		return nil, err
	}
	logger.Info("NimbusPolicy configured", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)

	return nimbusPolicy, nil
}

func (r *SecurityIntentBindingReconciler) updateStatus(ctx context.Context, logger logr.Logger, req ctrl.Request) error {
	// To handle potential latency issues with the Kubernetes API server, we
	// implement an exponential backoff strategy when fetching the NimbusPolicy
	// custom resource. This enhances resilience by retrying failed requests with
	// increasing intervals, preventing excessive retries in case of persistent 'Not
	// Found' errors.
	if retryErr := retry.OnError(retry.DefaultRetry, apierrors.IsNotFound, func() error {
		np := &v1.NimbusPolicy{}
		if err := r.Get(ctx, req.NamespacedName, np); err != nil {
			return err
		}
		return nil
	}); retryErr != nil {
		logger.Error(retryErr, "failed to fetch NimbusPolicy", "NimbusPolicy.Name", req.Name, "NimbusPolicy.Namespace", req.Namespace)
		return retryErr
	}

	// Since multiple adapters may update the NimbusPolicy status concurrently,
	// there's a risk of conflict during updates of NimbusPolicy status. To ensure
	// data consistency, retry on write failures. On conflict, the status update is
	// retried with an exponential backoff strategy. This provides resilience against
	// potential issues while preventing indefinite retries in case of persistent
	// conflicts.
	if retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestNp := &v1.NimbusPolicy{}
		if err := r.Get(ctx, req.NamespacedName, latestNp); err != nil {
			return err
		}

		latestNp.Status = v1.NimbusPolicyStatus{
			Status:      StatusCreated,
			LastUpdated: metav1.Now(),
		}
		if err := r.Status().Update(ctx, latestNp); err != nil {
			return err
		}
		return nil
	}); retryErr != nil {
		logger.Error(retryErr, "failed to update NimbusPolicy status", "NimbusPolicy.Name", req.Name, "NimbusPolicy.Namespace", req.Namespace)
		return retryErr
	}

	// Fetch the latest SecurityIntentBinding so that we have the latest state
	// on the cluster.
	latestSib := &v1.SecurityIntentBinding{}
	if err := r.Get(ctx, req.NamespacedName, latestSib); err != nil {
		logger.Error(err, "failed to fetch SecurityIntentBinding", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
		return err
	}
	count, boundIntents := extractBoundIntentsInfo(latestSib.Spec.Intents)
	latestSib.Status = v1.SecurityIntentBindingStatus{
		Status:               StatusCreated,
		LastUpdated:          metav1.Now(),
		NumberOfBoundIntents: count,
		BoundIntents:         boundIntents,
		NimbusPolicy:         req.Name,
	}
	if err := r.Status().Update(ctx, latestSib); err != nil {
		logger.Error(err, "failed to update SecurityIntentBinding status", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
		return err
	}

	return nil
}
