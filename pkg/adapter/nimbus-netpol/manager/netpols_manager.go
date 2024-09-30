// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package manager

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	netv1 "k8s.io/api/networking/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/common"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	adapterutil "github.com/5GSEC/nimbus/pkg/adapter/util"
	globalwatcher "github.com/5GSEC/nimbus/pkg/adapter/watcher"

	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-netpol/processor"
	netpolwatcher "github.com/5GSEC/nimbus/pkg/adapter/nimbus-netpol/watcher"
)

var (
	scheme    = runtime.NewScheme()
	k8sClient client.Client
)

func init() {
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(netv1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	k8sClient = k8s.NewOrDie(scheme)
}

func Run(ctx context.Context) {

	// Watch NimbusPolicies only, and not ClusterNimbusPolicies as NetworkPolicy is
	// namespaced scoped
	npCh := make(chan common.Request)
	deletedNpCh := make(chan *unstructured.Unstructured)
	go globalwatcher.WatchNimbusPolicies(ctx, npCh, deletedNpCh, "SecurityIntentBinding", "ClusterSecurityIntentBinding")

	updatedNetpolCh := make(chan common.Request)
	deletedNetpolCh := make(chan common.Request)
	go netpolwatcher.WatchNetpols(ctx, updatedNetpolCh, deletedNetpolCh)

	for {
		select {
		case <-ctx.Done():
			close(npCh)
			close(deletedNpCh)
			close(updatedNetpolCh)
			close(deletedNetpolCh)
			return
		case createdNp := <-npCh:
			createOrUpdateNetworkPolicy(ctx, createdNp.Name, createdNp.Namespace)
		case deletedNp := <-deletedNpCh:
			logNetworkPolicyToDelete(ctx, deletedNp)
		case updatedNetpol := <-updatedNetpolCh:
			reconcileNetPol(ctx, updatedNetpol.Name, updatedNetpol.Namespace, false)
		case deletedNetpol := <-deletedNetpolCh:
			reconcileNetPol(ctx, deletedNetpol.Name, deletedNetpol.Namespace, true)
		}
	}
}

func reconcileNetPol(ctx context.Context, netpolName, namespace string, deleted bool) {
	logger := log.FromContext(ctx)
	npName := adapterutil.ExtractAnyNimbusPolicyName(netpolName)
	var np v1alpha1.NimbusPolicy
	err := k8sClient.Get(ctx, types.NamespacedName{Name: npName, Namespace: namespace}, &np)
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "failed to get NimbusPolicy", "NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", namespace)
		}
		return
	}
	if deleted {
		logger.V(2).Info("Reconciling deleted NetworkPolicy", "NetworkPolicy.Name", netpolName, "NetworkPolicy.Namespace", namespace)
	} else {
		logger.V(2).Info("Reconciling modified NetworkPolicy", "NetworkPolicy.Name", netpolName, "NetworkPolicy.Namespace", namespace)
	}
	createOrUpdateNetworkPolicy(ctx, npName, namespace)
}

func createOrUpdateNetworkPolicy(ctx context.Context, npName, npNamespace string) {
	logger := log.FromContext(ctx)
	var np v1alpha1.NimbusPolicy
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: npName, Namespace: npNamespace}, &np); err != nil {
		logger.Error(err, "failed to get NimbusPolicy", "NimbusPolicy.Name", npName[0], "NimbusPolicy.Namespace", npName[1])
		return
	}

	if adapterutil.IsOrphan(np.GetOwnerReferences(), "SecurityIntentBinding", "ClusterSecurityIntentBinding") {
		logger.V(4).Info("Ignoring orphan NimbusPolicy", "NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", npNamespace)
		return
	}

	deleteDanglingNetpols(ctx, np, logger)
	netPols := processor.BuildNetPolsFrom(logger, np, k8sClient)
	// Iterate using a separate index variable to avoid aliasing
	for idx := range netPols {
		netpol := netPols[idx]

		// Set NimbusPolicy as the owner of the network policy
		if err := ctrl.SetControllerReference(&np, &netpol, scheme); err != nil {
			logger.Error(err, "failed to set OwnerReference on NetworkPolicy", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
			return
		}

		var existingNetpol netv1.NetworkPolicy
		err := k8sClient.Get(ctx, types.NamespacedName{Name: netpol.Name, Namespace: netpol.Namespace}, &existingNetpol)
		if err != nil && !errors.IsNotFound(err) {
			logger.Error(err, "failed to get existing NetworkPolicy", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
			return
		}
		if err != nil {
			if errors.IsNotFound(err) {
				if err = k8sClient.Create(ctx, &netpol); err != nil {
					logger.Error(err, "failed to create NetworkPolicy", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
					return
				}
				logger.Info("NetworkPolicy created", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
			}
		} else {
			netpol.ObjectMeta.ResourceVersion = existingNetpol.ObjectMeta.ResourceVersion
			if err = k8sClient.Update(ctx, &netpol); err != nil {
				logger.Error(err, "failed to configure existing NetworkPolicy", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
				return
			}
			logger.Info("NetworkPolicy configured", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
		}

		if err = adapterutil.UpdateNpStatus(ctx, k8sClient, "NetworkPolicy/"+netpol.Name, np.Name, np.Namespace, false); err != nil {
			logger.Error(err, "failed to update NetworkPolicies status in NimbusPolicy")
		}
	}
}

func logNetworkPolicyToDelete(ctx context.Context, deletedNp *unstructured.Unstructured) {
	logger := log.FromContext(ctx)

	var netpols netv1.NetworkPolicyList
	if err := k8sClient.List(ctx, &netpols, &client.ListOptions{Namespace: deletedNp.GetNamespace()}); err != nil {
		logger.Error(err, "failed to list NetworkPolicies")
		return
	}

	// Kubernetes GC automatically deletes the child when the parent/owner is
	// deleted. So, we don't need to delete the policy because NimbusPolicy is the
	// owner and when it gets deleted all the corresponding policies will be
	// automatically deleted.
	for _, netpol := range netpols.Items {
		for _, ownerRef := range netpol.OwnerReferences {
			if ownerRef.Name == deletedNp.GetName() && ownerRef.UID == deletedNp.GetUID() {
				logger.Info("NetworkPolicy already deleted due to NimbusPolicy deletion",
					"NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace,
					"NetworkPolicy.Name", deletedNp.GetName(), "NetworkPolicy.Namespace", deletedNp.GetNamespace(),
				)
				break
			}
		}
	}
}

func deleteDanglingNetpols(ctx context.Context, np v1alpha1.NimbusPolicy, logger logr.Logger) {
	var existingNetpols netv1.NetworkPolicyList
	if err := k8sClient.List(ctx, &existingNetpols, client.InNamespace(np.Namespace)); err != nil {
		logger.Error(err, "failed to list NetworkPolicies for cleanup")
		return
	}

	var netpolsOwnedByNp []netv1.NetworkPolicy
	for _, netpol := range existingNetpols.Items {
		for _, ownerRef := range netpol.OwnerReferences {
			if ownerRef.Name == np.Name && ownerRef.UID == np.UID {
				netpolsOwnedByNp = append(netpolsOwnedByNp, netpol)
				break
			}
		}
	}

	if len(netpolsOwnedByNp) == 0 {
		return
	}

	netpolsToDelete := make(map[string]netv1.NetworkPolicy)

	// Populate owned Netpols
	for _, netpolOwnedByNp := range netpolsOwnedByNp {
		netpolsToDelete[netpolOwnedByNp.Name] = netpolOwnedByNp
	}

	for _, nimbusRule := range np.Spec.NimbusRules {
		netpolName := np.Name + "-" + strings.ToLower(nimbusRule.ID)
		delete(netpolsToDelete, netpolName)
	}

	for netpolName := range netpolsToDelete {
		netpol := netpolsToDelete[netpolName]
		if err := k8sClient.Delete(ctx, &netpol); err != nil {
			logger.Error(err, "failed to delete dangling NetworkPolicy", "NetworkPolicy.Name", netpol.Namespace, "NetworkPolicy.Namespace", netpol.Namespace)
			continue
		}

		if err := adapterutil.UpdateNpStatus(ctx, k8sClient, "NetworkPolicy/"+netpol.Name, np.Name, np.Namespace, true); err != nil {
			logger.Error(err, "failed to update NetworkPolicy status in NimbusPolicy")
		}
		logger.Info("Dangling NetworkPolicy deleted", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
	}
}
