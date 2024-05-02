// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controller

import (
	"context"
	"errors"
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
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/5GSEC/nimbus/api/v1alpha"
	processorerrors "github.com/5GSEC/nimbus/pkg/processor/errors"
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
			return doNotRequeue()
		}
		logger.Error(err, "failed to fetch SecurityIntentBinding", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
		return requeueWithError(err)
	}

	if sib.GetGeneration() == 1 {
		logger.Info("SecurityIntentBinding found", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
	} else {
		logger.Info("SecurityIntentBinding configured", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
	}

	if err = r.updateSibStatus(ctx, logger, req); err != nil {
		return requeueWithError(err)
	}

	if err = r.createOrUpdateNp(ctx, logger, req); err != nil {
		return requeueWithError(err)
	}

	if err = r.updateSibStatusWithBoundSisAndNpInfo(ctx, logger, req); err != nil {
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
		Watches(&v1.SecurityIntent{},
			handler.EnqueueRequestsFromMapFunc(r.findSibsForSi),
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
	if _, ok := obj.(*v1.SecurityIntent); ok {
		return true
	}
	return ownerExists(r.Client, obj)
}

func (r *SecurityIntentBindingReconciler) createOrUpdateNp(ctx context.Context, logger logr.Logger, req ctrl.Request) error {
	// Always fetch the CRs so that we have the latest state of the CRs on the
	// cluster.

	var sib v1.SecurityIntentBinding
	if err := r.Get(ctx, req.NamespacedName, &sib); err != nil {
		logger.Error(err, "failed to fetch SecurityIntentBinding", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
		return err
	}

	var np v1.NimbusPolicy
	err := r.Get(ctx, req.NamespacedName, &np)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.createNp(ctx, logger, sib)
		}
		logger.Error(err, "failed to fetch NimbusPolicy", "NimbusPolicy.Name", req.Name, "NimbusPolicy.Namespace", req.Namespace)
		return err
	}
	return r.updateNp(ctx, logger, sib)
}

func (r *SecurityIntentBindingReconciler) createNp(ctx context.Context, logger logr.Logger, sib v1.SecurityIntentBinding) error {
	nimbusPolicy, err := policybuilder.BuildNimbusPolicy(ctx, logger, r.Client, r.Scheme, sib)
	// TODO: Improve error handling for CEL
	if err != nil {
		// Error is caused due to CEL, so don't retry to build NimbusPolicy.
		if strings.Contains(err.Error(), "error processing CEL") {
			logger.Error(err, "failed to build NimbusPolicy")
			return nil
		}
		if errors.Is(err, processorerrors.ErrSecurityIntentsNotFound) {
			// Since the SecurityIntent(s) referenced in SecurityIntentBinding spec do not
			// exist, so delete NimbusPolicy if it exists.
			if err := r.deleteNp(ctx, sib.GetName(), sib.GetNamespace()); err != nil {
				return err
			}
			return nil
		}
		logger.Error(err, "failed to build NimbusPolicy")
		return err
	}
	if nimbusPolicy == nil {
		logger.Info("Abort NimbusPolicy creation as no labels matched the CEL expressions")
		return nil
	}

	if err := r.Create(ctx, nimbusPolicy); err != nil {
		logger.Error(err, "failed to create NimbusPolicy", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)
		return err
	}
	logger.Info("NimbusPolicy created", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)

	return r.updateNpStatus(ctx, logger, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: sib.Namespace,
			Name:      sib.Name,
		}},
	)
}

func (r *SecurityIntentBindingReconciler) updateNp(ctx context.Context, logger logr.Logger, sib v1.SecurityIntentBinding) error {
	var existingNp v1.NimbusPolicy
	if err := r.Get(ctx, types.NamespacedName{Name: sib.Name, Namespace: sib.Namespace}, &existingNp); err != nil {
		logger.Error(err, "failed to fetch NimbusPolicy", "NimbusPolicy.Name", sib.Name, "NimbusPolicy.Namespace", sib.Namespace)
		return err
	}

	nimbusPolicy, err := policybuilder.BuildNimbusPolicy(ctx, logger, r.Client, r.Scheme, sib)
	// TODO: Improve error handling for CEL
	if err != nil {
		// Error is caused due to CEL, so don't retry to build NimbusPolicy.
		if strings.Contains(err.Error(), "error processing CEL") {
			logger.Error(err, "failed to build NimbusPolicy")
			return nil
		}
		if errors.Is(err, processorerrors.ErrSecurityIntentsNotFound) {
			// Since the SecurityIntent(s) referenced in SecurityIntentBinding spec do not
			// exist, so delete NimbusPolicy if it exists.
			if err := r.deleteNp(ctx, sib.GetName(), sib.GetNamespace()); err != nil {
				return err
			}
			return nil
		}
		logger.Error(err, "failed to build NimbusPolicy")
		return err
	}
	if nimbusPolicy == nil {
		logger.Info("Abort NimbusPolicy creation as no labels matched the CEL expressions")
		return nil
	}

	nimbusPolicy.ObjectMeta.ResourceVersion = existingNp.ObjectMeta.ResourceVersion
	if err := r.Update(ctx, nimbusPolicy); err != nil {
		logger.Error(err, "failed to configure NimbusPolicy", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)
		return err
	}
	logger.Info("NimbusPolicy configured", "NimbusPolicy.Name", nimbusPolicy.Name, "NimbusPolicy.Namespace", nimbusPolicy.Namespace)

	return r.updateNpStatus(ctx, logger, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: sib.Namespace,
			Name:      sib.Name,
		}},
	)
}

func (r *SecurityIntentBindingReconciler) findSibsForSi(ctx context.Context, si client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	sibs := &v1.SecurityIntentBindingList{}
	if err := r.List(ctx, sibs); err != nil {
		logger.Error(err, "failed to list SecurityIntentBindings")
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(sibs.Items))

	for idx, sib := range sibs.Items {
		for _, intent := range sib.Spec.Intents {
			if intent.Name == si.GetName() {
				requests[idx] = ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: sib.GetNamespace(),
						Name:      sib.GetName(),
					},
				}
				break
			}
		}
	}

	return requests
}

func (r *SecurityIntentBindingReconciler) deleteNp(ctx context.Context, name, namespace string) error {
	logger := log.FromContext(ctx)

	var np v1.NimbusPolicy
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &np)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	logger.Info("Deleting NimbusPolicy since no SecurityIntents found", "nimbusPolicyName", name, "nimbusPolicyNamespace", namespace)
	logger.Info("NimbusPolicy deleted", "nimbusPolicyName", name, "nimbusPolicyNamespace", namespace)
	if err = r.Delete(context.Background(), &np); err != nil {
		logger.Error(err, "failed to delete NimbusPolicy", "nimbusPolicyName", name, "nimbusPolicyNamespace", namespace)
		return err
	}

	return nil
}

func (r *SecurityIntentBindingReconciler) updateNpStatus(ctx context.Context, logger logr.Logger, req ctrl.Request) error {
	np := &v1.NimbusPolicy{}

	// To handle potential latency or outdated cache issues with the Kubernetes API
	// server, we implement an exponential backoff strategy when fetching the
	// NimbusPolicy custom resource. This enhances resilience by retrying failed
	// requests with increasing intervals, preventing excessive retries in case of
	// persistent 'Not Found' errors.
	if retryErr := retry.OnError(retry.DefaultRetry, apierrors.IsNotFound, func() error {
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
		if err := r.Get(ctx, req.NamespacedName, np); err != nil {
			return err
		}

		np.Status.Status = StatusCreated
		np.Status.LastUpdated = metav1.Now()
		if err := r.Status().Update(ctx, np); err != nil {
			return err
		}
		return nil
	}); retryErr != nil {
		logger.Error(retryErr, "failed to update NimbusPolicy status", "NimbusPolicy.Name", req.Name, "NimbusPolicy.Namespace", req.Namespace)
		return retryErr
	}
	return nil
}

func (r *SecurityIntentBindingReconciler) updateSibStatus(ctx context.Context, logger logr.Logger, req ctrl.Request) error {
	if retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestSib := &v1.SecurityIntentBinding{}
		if err := r.Get(ctx, req.NamespacedName, latestSib); err != nil {
			logger.Error(err, "failed to fetch SecurityIntentBinding", "securityIntentBindingName", req.Name, "securityIntentBindingNamespace", req.Namespace)
			return err
		}

		latestSib.Status.Status = StatusCreated
		latestSib.Status.LastUpdated = metav1.Now()

		if err := r.Status().Update(ctx, latestSib); err != nil {
			return err
		}
		return nil
	}); retryErr != nil {
		logger.Error(retryErr, "failed to update SecurityIntentBinding status", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
		return retryErr
	}

	return nil
}

func (r *SecurityIntentBindingReconciler) updateSibStatusWithBoundSisAndNpInfo(ctx context.Context, logger logr.Logger, req ctrl.Request) error {
	latestSib := &v1.SecurityIntentBinding{}
	if err := r.Get(ctx, req.NamespacedName, latestSib); err != nil {
		logger.Error(err, "failed to fetch SecurityIntentBinding", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
		return err
	}

	latestNp := &v1.NimbusPolicy{}
	if retryErr := retry.OnError(retry.DefaultRetry, apierrors.IsNotFound, func() error {
		if err := r.Get(ctx, req.NamespacedName, latestNp); err != nil {
			return err
		}
		return nil
	}); retryErr != nil {
		if !apierrors.IsNotFound(retryErr) {
			logger.Error(retryErr, "failed to fetch NimbusPolicy", "NimbusPolicy.Name", req.Name, "NimbusPolicy.Namespace", req.Namespace)
			return retryErr
		}

		// Remove outdated SecurityIntent(s) and NimbusPolicy info
		latestSib.Status.NumberOfBoundIntents = 0
		latestSib.Status.BoundIntents = nil
		latestSib.Status.NimbusPolicy = ""
		if err := r.Status().Update(ctx, latestSib); err != nil {
			logger.Error(err, "failed to update SecurityIntentBinding status", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
			return err
		}

		return nil
	}

	// Update SecurityIntentBinding status with bound SecurityIntent(s) and NimbusPolicy.
	latestSib.Status.NumberOfBoundIntents = int32(len(latestNp.Spec.NimbusRules))
	latestSib.Status.BoundIntents = extractBoundIntentsNameFromSib(ctx, r.Client, req.Name, req.Namespace)
	latestSib.Status.NimbusPolicy = req.Name

	if err := r.Status().Update(ctx, latestSib); err != nil {
		logger.Error(err, "failed to update SecurityIntentBinding status", "SecurityIntentBinding.Name", req.Name, "SecurityIntentBinding.Namespace", req.Namespace)
		return err
	}

	return nil
}
