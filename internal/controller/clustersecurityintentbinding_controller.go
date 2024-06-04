// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package controller

import (
	"context"
	"errors"
	"slices"

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

	v1alpha1 "github.com/5GSEC/nimbus/api/v1alpha1"
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
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterSecurityIntentBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	csib := &v1alpha1.ClusterSecurityIntentBinding{}
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

	// Check if the object was previouly marked as invalid
	if csib.Status.Status == StatusValidationFail {
		logger.Info("ClusterSecurityIntentBinding found, not valid", "ClusterSecurityIntentBinding.Name", req.Name)
		return doNotRequeue()
	}

	if csib.Status.Status == "" && !r.isValidCsib(ctx, logger, req) {
		if err = r.updateCsibStatus(ctx, logger, req, StatusValidationFail); err != nil {
			return requeueWithError(err)
		}
		return doNotRequeue()
	}

	if err = r.updateCsibStatus(ctx, logger, req, StatusCreated); err != nil {
		return requeueWithError(err)
	}

	if err = r.createOrUpdateCwnp(ctx, logger, req); err != nil {
		return requeueWithError(err)
	}

	if err = r.updateCSibStatusWithBoundSisAndCwnpInfo(ctx, logger, req); err != nil {
		return requeueWithError(err)
	}

	// Create the namespaced Nimbus policies
	if err = r.createOrUpdateNp(ctx, logger, req); err != nil {
		return requeueWithError(err)
	}

	return doNotRequeue()
}

// SetupWithManager sets up the controller with the Manager.
// WithEventFilter sets up the global predicates for a watch
func (r *ClusterSecurityIntentBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ClusterSecurityIntentBinding{}).
		Owns(&v1alpha1.ClusterNimbusPolicy{}).
		Owns(&v1alpha1.NimbusPolicy{}).
		WithEventFilter(
			predicate.Funcs{
				CreateFunc: r.createFn,
				UpdateFunc: r.updateFn,
				DeleteFunc: r.deleteFn,
			},
		).
		Watches(&v1alpha1.SecurityIntent{},
			handler.EnqueueRequestsFromMapFunc(r.findCsibsForSi),
		).
		Watches(&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(r.findCsibsForNamespace),
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					if e.ObjectNew.GetDeletionTimestamp() != nil {
						return true
					} else {
						return false
					}
				},
			}),
		).
		Complete(r)
}

func (r *ClusterSecurityIntentBindingReconciler) createFn(createEvent event.CreateEvent) bool {
	if _, ok := createEvent.Object.(*v1alpha1.ClusterNimbusPolicy); ok {
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
	if _, ok := obj.(*v1alpha1.ClusterSecurityIntentBinding); ok {
		return true
	}
	if _, ok := obj.(*v1alpha1.SecurityIntent); ok {
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
	var csib v1alpha1.ClusterSecurityIntentBinding
	if err := r.Get(ctx, req.NamespacedName, &csib); err != nil {
		logger.Error(err, "failed to fetch ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding.Name", req.Name)
		return err
	}

	var cwnp v1alpha1.ClusterNimbusPolicy
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

func (r *ClusterSecurityIntentBindingReconciler) createCwnp(ctx context.Context, logger logr.Logger, csib v1alpha1.ClusterSecurityIntentBinding) error {
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

func (r *ClusterSecurityIntentBindingReconciler) updateCwnp(ctx context.Context, logger logr.Logger, csib v1alpha1.ClusterSecurityIntentBinding) error {
	var existingCwnp v1alpha1.ClusterNimbusPolicy
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

	csibs := &v1alpha1.ClusterSecurityIntentBindingList{}
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

type npTrackingObj struct {
	create bool
	update bool
	np     *v1alpha1.NimbusPolicy
}

// we should not create object in these ns
var nsBlackList = []string{"kube-system"}

const wildcard = "*"

func (r *ClusterSecurityIntentBindingReconciler) isValidCsib(ctx context.Context, logger logr.Logger, req ctrl.Request) bool {

	// get the csib
	var csib v1alpha1.ClusterSecurityIntentBinding
	if err := r.Get(ctx, req.NamespacedName, &csib); err != nil {
		logger.Error(err, "failed to fetch ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding.Name", req.Name)
		return false
	}

	// validate the CSIB.
	excludeLen := len(csib.Spec.Selector.NsSelector.ExcludeNames)
	matchLen := len(csib.Spec.Selector.NsSelector.MatchNames)
	if matchLen > 0 && excludeLen > 0 {
		err := errors.New("invalid clustersecurityintentbinding")
		logger.Error(err, "Both MatchNames and ExcludeNames should not be set", "ClusterSecurityIntentBinding.Name", req.Name)
		return false
	}
	if matchLen == 0 && excludeLen == 0 {
		err := errors.New("invalid clustersecurityintentbinding")
		logger.Error(err, "Atleast one of MatchNames or ExcludeNames should be set", "ClusterSecurityIntentBinding.Name", req.Name)
		return false
	}
	// In MatchNames, if a  "*" is present, it should be the only entry
	for i, ns := range csib.Spec.Selector.NsSelector.MatchNames {
		if ns == wildcard && i > 0 {
			err := errors.New("invalid clustersecurityintentbinding")
			logger.Error(err, "If * is present, it should be only entry", "ClusterSecurityIntentBinding.Name", req.Name)
			return false
		}
	}

	return true
}

func (r *ClusterSecurityIntentBindingReconciler) createOrUpdateNp(ctx context.Context, logger logr.Logger, req ctrl.Request) error {

	// Reconcile the Nimbus Policies with Security Intents, CSIB, NimbusPolicyList, Namespaces

	// get the csib
	var csib v1alpha1.ClusterSecurityIntentBinding
	if err := r.Get(ctx, req.NamespacedName, &csib); err != nil {
		logger.Error(err, "failed to fetch ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding.Name", req.Name)
		return err
	}

	// get the nimbus policies
	// TODO: we might want to index the nimbus policies based on the owner since we are anyways filtering
	// based on the owner later
	var npList v1alpha1.NimbusPolicyList
	err := r.List(ctx, &npList)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "failed to fetch list of NimbusPolicy", "ClusterNimbusPolicy.Name", req.Name)
		return err
	}

	// Populate the NP tracking list. Filter out nimbus policies which are owned by other CSIB/SIB
	var npFilteredTrackingList []npTrackingObj
	for _, np := range npList.Items {
		for _, ref := range np.ObjectMeta.OwnerReferences {
			if csib.ObjectMeta.UID == ref.UID {
				npFilteredTrackingList = append(npFilteredTrackingList, npTrackingObj{np: &np})
				break
			}
		}
	}

	var nsList corev1.NamespaceList
	err = r.List(ctx, &nsList)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "failed to fetch list of Namespaces", "ClusterNimbusPolicy.Name", req.Name)
		return err
	}

	// Populate a map with all namespaces
	nsMap := make(map[string]corev1.Namespace)
	for _, nso := range nsList.Items {
		nsMap[nso.Name] = nso
	}

	// filter out the blacklist, deleted namespaces
	for ns, nsObj := range nsMap {
		if slices.Contains(nsBlackList, ns) {
			delete(nsMap, ns)
			continue
		}

		if nsObj.GetDeletionTimestamp() != nil {
			delete(nsMap, ns)
			continue
		}

		if len(csib.Spec.Selector.NsSelector.ExcludeNames) > 0 {
			if slices.Contains(csib.Spec.Selector.NsSelector.ExcludeNames, ns) {
				delete(nsMap, ns)
				continue
			}
		} else if ml := len(csib.Spec.Selector.NsSelector.MatchNames); ml > 0 {
			if ml == 1 && csib.Spec.Selector.NsSelector.MatchNames[0] == wildcard {
				continue
			}
			if !slices.Contains(csib.Spec.Selector.NsSelector.MatchNames, ns) {
				delete(nsMap, ns)
				continue
			}
		}
	}

	// The nsMap is the spec. We need to ensure that there are NP
	// for the specified namespaces. 3 cases here
	//   - If a namespace is in spec, and in the NP list, then mark NP for update.
	//   - If the namespace is in spec, and not there in the NP list, then
	//     build an NP for this namespace, and mark it for create.
	//   - If there are NPs in namespaces which are not in the spec list, delete
	//     those NPs
	for _, nsSpec := range nsMap {
		var seen bool = false
		for index, np_actual := range npFilteredTrackingList {
			if nsSpec.Name == np_actual.np.Namespace {
				npFilteredTrackingList[index].update = true
				seen = true
				break
			}
		}
		if !seen {
			// construct the nimbus policy object as it is not present in cluster
			nimbusPolicy, err := policybuilder.BuildNimbusPolicyFromClusterBinding(ctx, logger, r.Client, r.Scheme, csib, nsSpec.Name)
			if err == nil {
				npFilteredTrackingList = append(npFilteredTrackingList, npTrackingObj{create: true, np: nimbusPolicy})
			}
		}
	}

	// run through the tracking list, and create/update/delete the nimbus policies
	for _, nobj := range npFilteredTrackingList {
		if nobj.create {
			if err := r.Create(ctx, nobj.np); err != nil {
				logger.Error(err, "failed to create NimbusPolicy", "NimbusPolicy.Name", nobj.np.Name)
				return err
			}
			npReq := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: nobj.np.GetNamespace(),
					Name:      nobj.np.GetName(),
				}}
			if err = r.updateNpStatus(ctx, logger, npReq, StatusCreated); err != nil {
				return err
			}
			logger.Info("NimbusPolicy created", "NimbusPolicy.Name", nobj.np.Name)

		} else if nobj.update {
			// update intents, parameters. Build a new Nimbus Policy
			// TODO: Might be more efficient to simply update the intents, params
			newNimbusPolicy, err := policybuilder.BuildNimbusPolicyFromClusterBinding(ctx, logger, r.Client, r.Scheme, csib, nobj.np.Namespace)
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

			// Check equality
			// Another option is to check which CSIB was used to generate this nimbus policy
			if reason, equal := nobj.np.Equal(*newNimbusPolicy); equal {
				logger.Info("NimbusPolicy not updated as objects are same", "NimbusPolicy.name", nobj.np.Name, "Namespace", nobj.np.Namespace)
				continue
			} else {
				logger.Info("NimbusPolicy updated as objects are not same", "NimbusPolicy.name", nobj.np.Name, "Namespace", nobj.np.Namespace, "Reason", reason)
			}

			newNimbusPolicy.ObjectMeta.ResourceVersion = nobj.np.ObjectMeta.ResourceVersion
			if err := r.Update(ctx, newNimbusPolicy); err != nil {
				logger.Error(err, "failed to update NimbusPolicy", "NimbusPolicy.Name", newNimbusPolicy.Name)
				return err
			}
			npReq := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: newNimbusPolicy.GetNamespace(),
					Name:      newNimbusPolicy.GetName(),
				}}
			if err = r.updateNpStatus(ctx, logger, npReq, StatusCreated); err != nil {
				return err
			}
			logger.Info("NimbusPolicy updated", "NimbusPolicy.Name", newNimbusPolicy.Name)

			return nil

		} else {
			// delete the object
			logger.Info("Deleting NimbusPolicy since no namespaces found", "NimbusPolicyName", nobj.np.Name)
			if err = r.Delete(ctx, nobj.np); err != nil {
				logger.Error(err, "failed to delete NimbusPolicy", "NimbusPolicyName", nobj.np.Name)
				return err
			}
			logger.Info("NimbusPolicy deleted", "NimbusPolicyName", nobj.np.Name)
		}
	}

	return nil
}

func (r *ClusterSecurityIntentBindingReconciler) updateNpStatus(ctx context.Context, logger logr.Logger, req ctrl.Request, status string) error {
	np := &v1alpha1.NimbusPolicy{}

	// Get the np object. This might take multiple retries since object might have been just created
	if retryErr := retry.OnError(retry.DefaultRetry, apierrors.IsNotFound, func() error {
		if err := r.Get(ctx, req.NamespacedName, np); err != nil {
			return err
		}
		return nil
	}); retryErr != nil {
		logger.Error(retryErr, "failed to fetch NimbusPolicy", "NimbusPolicy.Name", req.Name)
		return retryErr
	}

	if retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := r.Get(ctx, req.NamespacedName, np); err != nil {
			return err
		}

		np.Status.Status = status
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

func (r *ClusterSecurityIntentBindingReconciler) findCsibsForNamespace(ctx context.Context, nsObj client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	csibs := &v1alpha1.ClusterSecurityIntentBindingList{}
	if err := r.List(ctx, csibs); err != nil {
		logger.Error(err, "failed to list ClusterSecurityIntentBindings")
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(csibs.Items))

	for _, csib := range csibs.Items {

		var toBeReconciled = false

		if csib.Status.Status == StatusValidationFail {
			continue
		}

		/*
		 * If the csib has a wildcard, then it is of interest since
		 * we have to modify the number of fanout of the csib.
		 * In case of add, the fanout will increase, and in case of
		 * delete the fanout will reduce.
		 */
		if len(csib.Spec.Selector.NsSelector.MatchNames) == 1 &&
			csib.Spec.Selector.NsSelector.MatchNames[0] == wildcard {
			toBeReconciled = true
		} else if len(csib.Spec.Selector.NsSelector.MatchNames) > 0 {
			/*
			 * If the ns being added/deleted appears in the csib matchNames, then
			 * the csib is of interest
			 */
			if slices.Contains(csib.Spec.Selector.NsSelector.MatchNames, nsObj.GetName()) {
				toBeReconciled = true
			}
		}

		/*
		 * We need to reconcile if the namespace object does not appear
		 * in the CSIB exclude list
		 * For example, there was a excludeName consisting of ns_1, ns_2.
		 * and now ns_3 is added in the cluster. So, as part of
		 * reconciliation we now have to create NimbusPolicy for ns_3.
		 */
		if len(csib.Spec.Selector.NsSelector.ExcludeNames) > 0 {
			if !slices.Contains(csib.Spec.Selector.NsSelector.ExcludeNames, nsObj.GetName()) {
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

func (r *ClusterSecurityIntentBindingReconciler) updateCsibStatus(ctx context.Context, logger logr.Logger, req ctrl.Request, status string) error {
	if retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestCsib := &v1alpha1.ClusterSecurityIntentBinding{}
		if err := r.Get(ctx, req.NamespacedName, latestCsib); err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, "failed to fetch ClusterSecurityIntentBinding", "clusterSecurityIntentBindingName", req.Name)
			return err
		}

		latestCsib.Status.Status = status
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
	cwnp := &v1alpha1.ClusterNimbusPolicy{}

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

	var cwnp v1alpha1.ClusterNimbusPolicy
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
	latestCsib := &v1alpha1.ClusterSecurityIntentBinding{}
	if err := r.Get(ctx, req.NamespacedName, latestCsib); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "failed to fetch ClusterSecurityIntentBinding", "ClusterSecurityIntentBinding.Name", req.Name)
		return err
	}

	latestCwnp := &v1alpha1.ClusterNimbusPolicy{}
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
