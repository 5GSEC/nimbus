// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controller

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
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
			return doNotRequeue()
		}
		logger.Error(err, "failed to get ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding.Name", csib.Name)
		return requeueWithError(err)
	}

	if csib.GetGeneration() == 1 {
		logger.Info("ClusterSecurityIntentBinding found", "ClusterSecurityIntentBinding.Name", req.Name)
	} else {
		logger.Info("ClusterSecurityIntentBinding configured", "ClusterSecurityIntentBinding.Name", req.Name)
	}

	if err = r.updateCsibStatus(ctx, logger, req); err != nil {
		return requeueWithError(err)
	}

	if err = r.createOrUpdateCwnp(ctx, logger, req); err != nil {
		return requeueWithError(err)
	}

	if err = r.updateCSibStatusWithBoundSisAndCwnpInfo(ctx, logger, req); err != nil {
		return requeueWithError(err)
	}

	return doNotRequeue()
}

// SetupWithManager sets up the controller with the Manager.
// WithEventFilter sets up the global predicates for a watch
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
		Watches(&v1.SecurityIntent{},
			handler.EnqueueRequestsFromMapFunc(r.findCsibsForSi),
		).
		Watches(&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(r.findCsibsForNamespace),
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					// Ignore updates
					return false
				},
			}),
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
	if _, ok := obj.(*v1.SecurityIntent); ok {
		return true
	}
	if _, ok := obj.(*corev1.Namespace); ok {
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
		return err
	}
	return r.updateCwnp(ctx, logger, csib)
}

func (r *ClusterSecurityIntentBindingReconciler) createCwnp(ctx context.Context, logger logr.Logger, csib v1.ClusterSecurityIntentBinding) error {
	clusterNp, err := policybuilder.BuildClusterNimbusPolicy(ctx, logger, r.Client, r.Scheme, csib)
	if err != nil {
		if errors.Is(err, processorerrors.ErrSecurityIntentsNotFound) {
			// Since the SecurityIntent(s) referenced in ClusterSecurityIntentBinding spec do not
			// exist, so delete ClusterNimbusPolicy if it exists.
			if err := r.deleteCwnp(ctx, csib.GetName()); err != nil {
				return err
			}
			return nil
		}
		logger.Error(err, "failed to build ClusterNimbusPolicy")
		return err
	}

	if err := r.Create(ctx, clusterNp); err != nil {
		logger.Error(err, "failed to create ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", clusterNp.Name)
		return err
	}
	logger.Info("ClusterNimbusPolicy created", "ClusterNimbusPolicy.Name", clusterNp.Name)

	return r.updateCwnpStatus(ctx, logger, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: csib.Name,
		}},
	)
}

func (r *ClusterSecurityIntentBindingReconciler) updateCwnp(ctx context.Context, logger logr.Logger, csib v1.ClusterSecurityIntentBinding) error {
	var existingCwnp v1.ClusterNimbusPolicy
	if err := r.Get(ctx, types.NamespacedName{Name: csib.Name}, &existingCwnp); err != nil {
		logger.Error(err, "failed to fetch ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", csib.Name)
		return err
	}

	clusterNp, err := policybuilder.BuildClusterNimbusPolicy(ctx, logger, r.Client, r.Scheme, csib)
	if err != nil {
		if errors.Is(err, processorerrors.ErrSecurityIntentsNotFound) {
			// Since the SecurityIntent(s) referenced in ClusterSecurityIntentBinding spec do not
			// exist, so delete ClusterNimbusPolicy if it exists.
			if err := r.deleteCwnp(ctx, csib.GetName()); err != nil {
				return err
			}
			return nil
		}
		logger.Error(err, "failed to build ClusterNimbusPolicy")
		return err
	}

	clusterNp.ObjectMeta.ResourceVersion = existingCwnp.ObjectMeta.ResourceVersion
	if err := r.Update(ctx, clusterNp); err != nil {
		logger.Error(err, "failed to configure ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", clusterNp.Name)
		return err
	}
	logger.Info("ClusterNimbusPolicy configured", "ClusterNimbusPolicy.Name", clusterNp.Name)

	return r.updateCwnpStatus(ctx, logger, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: csib.Name,
		}},
	)
}

func (r *ClusterSecurityIntentBindingReconciler) findCsibsForSi(ctx context.Context, si client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	csibs := &v1.ClusterSecurityIntentBindingList{}
	if err := r.List(ctx, csibs); err != nil {
		logger.Error(err, "failed to list ClusterSecurityIntentBindings")
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(csibs.Items))

	for idx, csib := range csibs.Items {
		for _, intent := range csib.Spec.Intents {
			if intent.Name == si.GetName() {
				requests[idx] = ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: csib.GetNamespace(),
						Name:      csib.GetName(),
					},
				}
				break
			}
		}
	}

	return requests
}

func (r *ClusterSecurityIntentBindingReconciler) findCsibsForNamespace(ctx context.Context, nsObj client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	csibs := &v1.ClusterSecurityIntentBindingList{}
	if err := r.List(ctx, csibs); err != nil {
		logger.Error(err, "failed to list ClusterSecurityIntentBindings")
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(csibs.Items))

	for _, csib := range csibs.Items {

		var toBeReconciled bool = false
		/*
		 * If matchnames, and excludenames is zero, then this csib
		 * of interest since we have to modify the number of fanout.
		 * In case of add, the fanout will increase, and in case of
		 * delete the fanout will reduced.
		 */
		if len(csib.Spec.Selector.NsSelector.MatchNames) == 0 &&
			len(csib.Spec.Selector.NsSelector.ExcludeNames) == 0 {
			toBeReconciled = true
		}

		/*
		 * If the ns being added/deleted appears in the matchNames, then
		 * we do not do anything since the NimbusPolicy would have been
		 * generated for ns in the matchNames, and there is no fanout to be done
		 */
		if len(csib.Spec.Selector.NsSelector.MatchNames) > 0 {
			continue
		}

		/*
		 * We need to reconcile if the namespace object does not appear
		 * in the CSIB exclude list
		 * For example, there was a excludeName consisting of ns_1, ns_2.
		 * and now ns_2 does not appear in the excludeNames. So, as part of
		 * reconciliation we now have to create NimbusPolicy for ns_2.
		 */
		if len(csib.Spec.Selector.NsSelector.ExcludeNames) > 0 {
			var outOfSet bool = true
			for _, ns := range csib.Spec.Selector.NsSelector.ExcludeNames {
				if ns == nsObj.GetName() {
					outOfSet = false
					break
				}
			}
			if outOfSet {
				toBeReconciled = true
			}
		}
		if toBeReconciled {

			requests = append(requests, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: csib.GetNamespace(),
					Name:      csib.GetName(),
				},
			})
		}
	}

	return requests
}

func (r *ClusterSecurityIntentBindingReconciler) updateCsibStatus(ctx context.Context, logger logr.Logger, req ctrl.Request) error {
	if retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestCsib := &v1.ClusterSecurityIntentBinding{}
		if err := r.Get(ctx, req.NamespacedName, latestCsib); err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, "failed to fetch ClusterSecurityIntentBinding", "clusterSecurityIntentBindingName", req.Name)
			return err
		}

		latestCsib.Status.Status = StatusCreated
		latestCsib.Status.LastUpdated = metav1.Now()
		if err := r.Status().Update(ctx, latestCsib); err != nil {
			return err
		}

		return nil
	}); retryErr != nil {
		logger.Error(retryErr, "failed to update ClusterSecurityIntentBinding status", "ClusterSecurityIntentBinding.Name", req.Name)
		return retryErr
	}

	return nil
}

func (r *ClusterSecurityIntentBindingReconciler) updateCwnpStatus(ctx context.Context, logger logr.Logger, req ctrl.Request) error {
	cwnp := &v1.ClusterNimbusPolicy{}

	// To handle potential latency or outdated cache issues with the Kubernetes API
	// server, we implement an exponential backoff strategy when fetching the
	// ClusterNimbusPolicy custom resource. This enhances resilience by retrying
	// failed requests with increasing intervals, preventing excessive retries in
	// case of persistent 'Not Found' errors.
	if retryErr := retry.OnError(retry.DefaultRetry, apierrors.IsNotFound, func() error {
		if err := r.Get(ctx, req.NamespacedName, cwnp); err != nil {
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
		if err := r.Get(ctx, req.NamespacedName, cwnp); err != nil {
			return err
		}
		cwnp.Status.Status = StatusCreated
		cwnp.Status.LastUpdated = metav1.Now()
		if err := r.Status().Update(ctx, cwnp); err != nil {
			return err
		}
		return nil
	}); retryErr != nil {
		logger.Error(retryErr, "failed to update ClusterNimbusPolicy status", "ClusterNimbusPolicy.Name", req.Name)
		return retryErr
	}

	return nil
}

func (r *ClusterSecurityIntentBindingReconciler) deleteCwnp(ctx context.Context, name string) error {
	logger := log.FromContext(ctx)

	var cwnp v1.ClusterNimbusPolicy
	err := r.Get(ctx, types.NamespacedName{Name: name}, &cwnp)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	logger.Info("Deleting ClusterNimbusPolicy since no SecurityIntents found", "clusterNimbusPolicyName", name)
	logger.Info("ClusterNimbusPolicy deleted", "clusterNimbusPolicyName", name)
	if err = r.Delete(context.Background(), &cwnp); err != nil {
		logger.Error(err, "failed to delete ClusterNimbusPolicy", "clusterNimbusPolicyName", name)
		return err
	}

	return nil
}

func (r *ClusterSecurityIntentBindingReconciler) updateCSibStatusWithBoundSisAndCwnpInfo(ctx context.Context, logger logr.Logger, req ctrl.Request) error {
	latestCsib := &v1.ClusterSecurityIntentBinding{}
	if err := r.Get(ctx, req.NamespacedName, latestCsib); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "failed to fetch ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding.Name", req.Name)
		return err
	}

	latestCwnp := &v1.ClusterNimbusPolicy{}
	if retryErr := retry.OnError(retry.DefaultRetry, apierrors.IsNotFound, func() error {
		if err := r.Get(ctx, req.NamespacedName, latestCwnp); err != nil {
			return err
		}
		return nil
	}); retryErr != nil {
		if !apierrors.IsNotFound(retryErr) {
			logger.Error(retryErr, "failed to fetch ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", req.Name)
			return retryErr
		}

		// Remove outdated SecurityIntent(s) and ClusterNimbusPolicy info
		latestCsib.Status.NumberOfBoundIntents = 0
		latestCsib.Status.BoundIntents = nil
		latestCsib.Status.ClusterNimbusPolicy = ""
		if err := r.Status().Update(ctx, latestCsib); err != nil {
			logger.Error(err, "failed to update ClusterSecurityIntentBinding status", "ClusterSecurityIntentBinding.Name", latestCsib.Name)
			return err
		}
		return nil
	}

	// Update ClusterSecurityIntentBinding status with bound SecurityIntent(s) and NimbusPolicy.
	latestCsib.Status.NumberOfBoundIntents = int32(len(latestCwnp.Spec.NimbusRules))
	latestCsib.Status.BoundIntents = extractBoundIntentsNameFromCSib(ctx, r.Client, req.Name)
	latestCsib.Status.ClusterNimbusPolicy = req.Name

	if err := r.Status().Update(ctx, latestCsib); err != nil {
		logger.Error(err, "failed to update ClusterSecurityIntentBinding status", "ClusterSecurityIntentBinding.Name", latestCsib.Name)
		return err
	}

	return nil
}
