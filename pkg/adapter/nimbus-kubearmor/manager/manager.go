// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package manager

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
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

	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-kubearmor/processor"
	kspwatcher "github.com/5GSEC/nimbus/pkg/adapter/nimbus-kubearmor/watcher"
)

var (
	scheme    = runtime.NewScheme()
	k8sClient client.Client
)

func init() {
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(kubearmorv1.AddToScheme(scheme))
	k8sClient = k8s.NewOrDie(scheme)
}

func Run(ctx context.Context) {
	npCh := make(chan common.Request)
	deletedNpCh := make(chan *unstructured.Unstructured)
	go globalwatcher.WatchNimbusPolicies(ctx, npCh, deletedNpCh, "SecurityIntentBinding", "ClusterSecurityIntentBinding")

	updatedKspCh := make(chan common.Request)
	deletedKspCh := make(chan common.Request)
	go kspwatcher.WatchKsps(ctx, updatedKspCh, deletedKspCh)

	for {
		select {
		case <-ctx.Done():
			close(npCh)
			close(deletedNpCh)
			close(updatedKspCh)
			close(deletedKspCh)
			return
		case createdNp := <-npCh:
			createOrUpdateKsp(ctx, createdNp.Name, createdNp.Namespace)
		case deletedNp := <-deletedNpCh:
			logKspToDelete(ctx, deletedNp)
		case updatedKsp := <-updatedKspCh:
			reconcileKsp(ctx, updatedKsp.Name, updatedKsp.Namespace, false)
		case deletedKsp := <-deletedKspCh:
			reconcileKsp(ctx, deletedKsp.Name, deletedKsp.Namespace, true)
		}
	}
}

func reconcileKsp(ctx context.Context, kspName, namespace string, deleted bool) {
	logger := log.FromContext(ctx)
	npName := adapterutil.ExtractAnyNimbusPolicyName(kspName)
	var np v1alpha1.NimbusPolicy
	err := k8sClient.Get(ctx, types.NamespacedName{Name: npName, Namespace: namespace}, &np)
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "failed to get NimbusPolicy", "NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", namespace)
		}
		return
	}
	if deleted {
		logger.V(2).Info("Reconciling deleted KubeArmorPolicy", "KubeArmorPolicy.Name", kspName, "KubeArmorPolicy.Namespace", namespace)
	} else {
		logger.V(2).Info("Reconciling modified KubeArmorPolicy", "KubeArmorPolicy.Name", kspName, "KubeArmorPolicy.Namespace", namespace)
	}
	createOrUpdateKsp(ctx, npName, namespace)
}

func createOrUpdateKsp(ctx context.Context, npName, npNamespace string) {
	logger := log.FromContext(ctx)
	var np v1alpha1.NimbusPolicy
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: npName, Namespace: npNamespace}, &np); err != nil {
		logger.Error(err, "failed to get NimbusPolicy", "NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", npNamespace)
		return
	}

	if adapterutil.IsOrphan(np.GetOwnerReferences(), "SecurityIntentBinding", "ClusterSecurityIntentBinding") {
		logger.V(4).Info("Ignoring orphan NimbusPolicy", "NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", npNamespace)
		return
	}

	deleteDanglingKsps(ctx, np, logger)
	ksps := processor.BuildKspsFrom(logger, &np)

	// Iterate using a separate index variable to avoid aliasing
	for idx := range ksps {
		ksp := ksps[idx]

		// Set NimbusPolicy as the owner of the KSP
		if err := ctrl.SetControllerReference(&np, &ksp, scheme); err != nil {
			logger.Error(err, "failed to set OwnerReference on KubeArmorPolicy", "Name", ksp.Name)
			return
		}

		var existingKsp kubearmorv1.KubeArmorPolicy
		err := k8sClient.Get(ctx, types.NamespacedName{Name: ksp.Name, Namespace: ksp.Namespace}, &existingKsp)
		if err != nil && !errors.IsNotFound(err) {
			logger.Error(err, "failed to get existing KubeArmorPolicy", "KubeArmorPolicy.Name", ksp.Name, "KubeArmorPolicy.Namespace", ksp.Namespace)
			return
		}
		if err != nil {
			if errors.IsNotFound(err) {
				if err = k8sClient.Create(ctx, &ksp); err != nil {
					logger.Error(err, "failed to create KubeArmorPolicy", "KubeArmorPolicy.Name", ksp.Name, "KubeArmorPolicy.Namespace", ksp.Namespace)
					return
				}
				logger.Info("KubeArmorPolicy created", "KubeArmorPolicy.Name", ksp.Name, "KubeArmorPolicy.Namespace", ksp.Namespace)
			}
		} else {
			ksp.ObjectMeta.ResourceVersion = existingKsp.ObjectMeta.ResourceVersion
			if err = k8sClient.Update(ctx, &ksp); err != nil {
				logger.Error(err, "failed to configure existing KubeArmorPolicy", "KubeArmorPolicy.Name", existingKsp.Name, "KubeArmorPolicy.Namespace", existingKsp.Namespace)
				return
			}
			logger.Info("KubeArmorPolicy configured", "KubeArmorPolicy.Name", existingKsp.Name, "KubeArmorPolicy.Namespace", existingKsp.Namespace)
		}

		if err = adapterutil.UpdateNpStatus(ctx, k8sClient, "KubeArmorPolicy/"+ksp.Name, np.Name, np.Namespace, false); err != nil {
			logger.Error(err, "failed to update KubeArmorPolicies status in NimbusPolicy")
		}
	}
}

func logKspToDelete(ctx context.Context, deletedNp *unstructured.Unstructured) {
	logger := log.FromContext(ctx)
	var ksps kubearmorv1.KubeArmorPolicyList

	if err := k8sClient.List(ctx, &ksps, &client.ListOptions{Namespace: deletedNp.GetNamespace()}); err != nil {
		logger.Error(err, "failed to list KubeArmorPolicies")
		return
	}

	// Kubernetes GC automatically deletes the child when the parent/owner is
	// deleted. So, we don't need to delete the policy because NimbusPolicy is the
	// owner and when it gets deleted all the corresponding policies will be
	// automatically deleted.
	for _, ksp := range ksps.Items {
		logger.Info("KubeArmorPolicy already deleted due to NimbusPolicy deletion",
			"KubeArmorPolicy.Name", ksp.Name, "KubeArmorPolicy.Namespace", ksp.Namespace,
			"NimbusPolicy.Name", deletedNp.GetName(), "NimbusPolicy.Namespace", deletedNp.GetNamespace(),
		)
	}
}

func deleteDanglingKsps(ctx context.Context, np v1alpha1.NimbusPolicy, logger logr.Logger) {
	var existingKsps kubearmorv1.KubeArmorPolicyList
	if err := k8sClient.List(ctx, &existingKsps, client.InNamespace(np.Namespace)); err != nil {
		logger.Error(err, "failed to list KubeArmorPolicies for cleanup")
		return
	}

	var kspsOwnedByNp []kubearmorv1.KubeArmorPolicy
	for _, ksp := range existingKsps.Items {
		for _, ownerRef := range ksp.OwnerReferences {
			if ownerRef.Name == np.Name && ownerRef.UID == np.UID {
				kspsOwnedByNp = append(kspsOwnedByNp, ksp)
				break
			}
		}
	}
	if len(kspsOwnedByNp) == 0 {
		return
	}

	kspsToDelete := make(map[string]kubearmorv1.KubeArmorPolicy)

	// Populate owned KSPs
	for _, kspOwnedByNp := range kspsOwnedByNp {
		kspsToDelete[kspOwnedByNp.Name] = kspOwnedByNp
	}

	for _, nimbusRule := range np.Spec.NimbusRules {
		kspName := np.Name + "-" + strings.ToLower(nimbusRule.ID)
		delete(kspsToDelete, kspName)
	}

	for kspName := range kspsToDelete {
		ksp := kspsToDelete[kspName]
		if err := k8sClient.Delete(ctx, &ksp); err != nil {
			logger.Error(err, "failed to delete dangling KubeArmorPolicy", "KubeArmorPolicy.Name", ksp.Namespace, "KubeArmorPolicy.Namespace", ksp.Namespace)
			continue
		}

		if err := adapterutil.UpdateNpStatus(ctx, k8sClient, "KubeArmorPolicy/"+ksp.Name, np.Name, np.Namespace, true); err != nil {
			logger.Error(err, "failed to update KubeArmorPolicy status in NimbusPolicy")
		}
		logger.Info("Dangling KubeArmorPolicy deleted", "KubeArmorPolicy.Name", ksp.Name, "KubeArmorPolicy.Namespace", ksp.Namespace)
	}
}
