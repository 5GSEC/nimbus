// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package manager

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	intentv1 "github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/common"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-kyverno/processor"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-kyverno/watcher"
	adapterutil "github.com/5GSEC/nimbus/pkg/adapter/util"
	globalwatcher "github.com/5GSEC/nimbus/pkg/adapter/watcher"
)

var (
	scheme    = runtime.NewScheme()
	k8sClient client.Client
)

func init() {
	utilruntime.Must(intentv1.AddToScheme(scheme))
	utilruntime.Must(kyvernov1.AddToScheme(scheme))
	k8sClient = k8s.NewOrDie(scheme)
}

func Run(ctx context.Context) {

	npCh := make(chan common.Request)
	deletedNpCh := make(chan common.Request)
	go globalwatcher.WatchNimbusPolicies(ctx, npCh, deletedNpCh, "SecurityIntentBinding")

	clusterNpChan := make(chan string)
	deletedClusterNpChan := make(chan string)
	go globalwatcher.WatchClusterNimbusPolicies(ctx, clusterNpChan, deletedClusterNpChan)

	updatedKcpCh := make(chan string)
	deletedKcpCh := make(chan string)

	go watcher.WatchKcps(ctx, updatedKcpCh, deletedKcpCh)

	updatedKpCh := make(chan common.Request)
	deletedKpCh := make(chan common.Request)

	go watcher.WatchKps(ctx, updatedKpCh, deletedKpCh)

	for {
		select {
		case <-ctx.Done():
			close(npCh)
			close(deletedNpCh)
			close(clusterNpChan)
			close(deletedClusterNpChan)
			close(updatedKcpCh)
			close(deletedKcpCh)
			close(updatedKpCh)
			close(deletedKpCh)
			return
		case createdNp := <-npCh:
			createOrUpdateKp(ctx, createdNp.Name, createdNp.Namespace)
		case createdCnp := <-clusterNpChan:
			createOrUpdateKcp(ctx, createdCnp)
		case deletedNp := <-deletedNpCh:
			deleteKp(ctx, deletedNp.Name, deletedNp.Namespace)
		case deletedCnp := <-deletedClusterNpChan:
			deleteKcp(ctx, deletedCnp)
		case updatedKp := <-updatedKpCh:
			reconcileKp(ctx, updatedKp.Name, updatedKp.Namespace, true)
		case updatedKcp := <-updatedKcpCh:
			reconcileKcp(ctx, updatedKcp, false)
		case deletedKcp := <-deletedKcpCh:
			reconcileKcp(ctx, deletedKcp, true)
		case deletedKp := <-deletedKpCh:
			reconcileKp(ctx, deletedKp.Name, deletedKp.Namespace, true)
		}

	}
}

func reconcileKp(ctx context.Context, kpName, namespace string, deleted bool) {
	logger := log.FromContext(ctx)
	npName := adapterutil.ExtractNpName(kpName)
	var np intentv1.NimbusPolicy
	err := k8sClient.Get(ctx, types.NamespacedName{Name: npName, Namespace: namespace}, &np)
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "failed to get NimbusPolicy", "NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", namespace)
		}
		return
	}
	if deleted {
		logger.V(2).Info("Reconciling deleted KyvernoPolicy", "KyvernoPolicy.Name", kpName, "KyvernoPolicy.Namespace", namespace)
	} else {
		logger.V(2).Info("Reconciling modified KyvernoPolicy", "KyvernoPolicy.Name", kpName, "KyvernoPolicy.Namespace", namespace)
	}
	createOrUpdateKp(ctx, npName, namespace)
}

func reconcileKcp(ctx context.Context, kcpName string, deleted bool) {
	logger := log.FromContext(ctx)
	cnpName := adapterutil.ExtractClusterNpName(kcpName)
	var cnp intentv1.ClusterNimbusPolicy
	err := k8sClient.Get(ctx, types.NamespacedName{Name: cnpName}, &cnp)
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "failed to get ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", cnpName)
		}
		return
	}
	if deleted {
		logger.V(2).Info("Reconciling deleted KyvernoClusterPolicy", "KyvernoClusterPolicy.Name", kcpName)
	} else {
		logger.V(2).Info("Reconciling modified KyvernoClusterPolicy", "KyvernoClusterPolicy.Name", kcpName)
	}
	createOrUpdateKcp(ctx, cnpName)
}

func createOrUpdateKp(ctx context.Context, npName, npNamespace string) {
	logger := log.FromContext(ctx)
	var np intentv1.NimbusPolicy
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: npName, Namespace: npNamespace}, &np); err != nil {
		logger.Error(err, "failed to get NimbusPolicy", "NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", npNamespace)
		return
	}

	if adapterutil.IsOrphan(np.GetOwnerReferences(), "SecurityIntentBinding") {
		logger.V(4).Info("Ignoring orphan NimbusPolicy", "NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", npNamespace)
		return
	}

	deleteDanglingkps(ctx, np, logger)
	kps := processor.BuildKpsFrom(logger, &np)

	// Iterate using a separate index variable to avoid aliasing
	for idx := range kps {
		kp := kps[idx]

		// Set NimbusPolicy as the owner of the KP
		if err := ctrl.SetControllerReference(&np, &kp, scheme); err != nil {
			logger.Error(err, "failed to set OwnerReference on KyvernoPolicy", "Name", kp.Name)
			return
		}

		var existingKp kyvernov1.Policy
		err := k8sClient.Get(ctx, types.NamespacedName{Name: kp.Name, Namespace: kp.Namespace}, &existingKp)
		if err != nil && !errors.IsNotFound(err) {
			logger.Error(err, "failed to get existing KyvernoPolicy", "KyvernoPolicy.Name", kp.Name, "KyvernoPolicy.Namespace", kp.Namespace)
			return
		}
		if err != nil {
			if errors.IsNotFound(err) {
				if err = k8sClient.Create(ctx, &kp); err != nil {
					logger.Error(err, "failed to create KyvernoPolicy", "KyvernoPolicy.Name", kp.Name, "KyvernoPolicy.Namespace", kp.Namespace)
					return
				}
				logger.Info("KyvernoPolicy created", "KyvernoPolicy.Name", kp.Name, "KyvernoPolicy.Namespace", kp.Namespace)
			}
		} else {
			kp.ObjectMeta.ResourceVersion = existingKp.ObjectMeta.ResourceVersion
			if err = k8sClient.Update(ctx, &kp); err != nil {
				logger.Error(err, "failed to configure existing KyvernoPolicy", "KyvernoPolicy.Name", existingKp.Name, "KyvernoPolicy.Namespace", existingKp.Namespace)
				return
			}
			logger.Info("KyvernoPolicy configured", "KyvernoPolicy.Name", existingKp.Name, "KyvernoPolicy.Namespace", existingKp.Namespace)
		}

		if err = adapterutil.UpdateNpStatus(ctx, k8sClient, "KyvernoPolicy/"+kp.Name, np.Name, np.Namespace, false); err != nil {
			logger.Error(err, "failed to update KyvernoPolicies status in NimbusPolicy")
		}
	}
}

func createOrUpdateKcp(ctx context.Context, cnpName string) {
	logger := log.FromContext(ctx)
	var cnp intentv1.ClusterNimbusPolicy
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: cnpName}, &cnp); err != nil {
		logger.Error(err, "failed to get ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", cnpName)
		return
	}

	if adapterutil.IsOrphan(cnp.GetOwnerReferences(), "ClusterSecurityIntentBinding") {
		logger.V(4).Info("Ignoring orphan ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", cnpName)
		return
	}

	deleteDanglingkcps(ctx, cnp, logger)
	kcps := processor.BuildKcpsFrom(logger, &cnp)

	for idx := range kcps {
		kcp := kcps[idx]

		// Set ClusterNimbusPolicy as the owner of the KCP
		if err := ctrl.SetControllerReference(&cnp, &kcp, scheme); err != nil {
			logger.Error(err, "failed to set OwnerReference on KyvernoClusterPolicy", "Name", kcp.Name)
			return
		}

		var existingKcp kyvernov1.ClusterPolicy
		err := k8sClient.Get(ctx, types.NamespacedName{Name: kcp.Name}, &existingKcp)
		if err != nil && !errors.IsNotFound(err) {
			logger.Error(err, "failed to get existing KyvernoClusterPolicy", "KyvernoClusterPolicy.Name", kcp.Name)
			return
		}
		if err != nil {
			if errors.IsNotFound(err) {
				if err = k8sClient.Create(ctx, &kcp); err != nil {
					logger.Error(err, "failed to create KyvernoClusterPolicy", "KyvernoClusterPolicy.Name", kcp.Name)
					return
				}
				logger.Info("KyvernoClusterPolicy created", "KyvernoClusterPolicy.Name", kcp.Name)
			}
		} else {
			kcp.ObjectMeta.ResourceVersion = existingKcp.ObjectMeta.ResourceVersion
			if err = k8sClient.Update(ctx, &kcp); err != nil {
				logger.Error(err, "failed to configure existing KyvernoClusterPolicy", "KyvernoClusterPolicy.Name", existingKcp.Name)
				return
			}
			logger.Info("KyvernoClusterPolicy configured", "KyvernoClusterPolicy.Name", existingKcp.Name)
		}

		if err = adapterutil.UpdateCnpStatus(ctx, k8sClient, "KyvernoClusterPolicy/"+kcp.Name, cnp.Name, false); err != nil {
			logger.Error(err, "failed to update KyvernoClusterPolicies status in NimbusPolicy")
		}
	}
}

func deleteKp(ctx context.Context, npName, npNamespace string) {
	logger := log.FromContext(ctx)
	var kps kyvernov1.PolicyList

	if err := k8sClient.List(ctx, &kps, &client.ListOptions{Namespace: npNamespace}); err != nil {
		logger.Error(err, "failed to list KyvernoPolicies")
		return
	}

	// Kubernetes GC automatically deletes the child when the parent/owner is
	// deleted. So, we don't need to do anything in this case since NimbusPolicy is
	// the owner and when it gets deleted corresponding kps will be automatically
	// deleted.
	for _, kp := range kps.Items {
		logger.Info("KyvernoPolicy already deleted due to NimbusPolicy deletion",
			"KyvernoPolicy.Name", kp.Name, "KyvernoPolicy.Namespace", kp.Namespace,
			"NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", npNamespace,
		)
	}
}

func deleteDanglingkps(ctx context.Context, np intentv1.NimbusPolicy, logger logr.Logger) {
	var existingkps kyvernov1.PolicyList
	if err := k8sClient.List(ctx, &existingkps, client.InNamespace(np.Namespace)); err != nil {
		logger.Error(err, "failed to list KyvernoPolicies for cleanup")
		return
	}

	var kpsOwnedByNp []kyvernov1.Policy
	for _, kp := range existingkps.Items {
		for _, ownerRef := range kp.OwnerReferences {
			if ownerRef.Name == np.Name && ownerRef.UID == np.UID {
				kpsOwnedByNp = append(kpsOwnedByNp, kp)
				break
			}
		}
	}
	if len(kpsOwnedByNp) == 0 {
		return
	}

	kpsToDelete := make(map[string]kyvernov1.Policy)

	// Populate owned kps
	for _, kpOwnedByNp := range kpsOwnedByNp {
		kpsToDelete[kpOwnedByNp.Name] = kpOwnedByNp
	}

	for _, nimbusRule := range np.Spec.NimbusRules {
		kpName := np.Name + "-" + strings.ToLower(nimbusRule.ID)
		delete(kpsToDelete, kpName)
	}

	for kpName := range kpsToDelete {
		kp := kpsToDelete[kpName]
		if err := k8sClient.Delete(ctx, &kp); err != nil {
			logger.Error(err, "failed to delete dangling KyvernoPolicy", "KyvernoPolicy.Name", kp.Namespace, "KyvernoPolicy.Namespace", kp.Namespace)
			continue
		}

		if err := adapterutil.UpdateNpStatus(ctx, k8sClient, "KyvernoPolicy/"+kp.Name, np.Name, np.Namespace, true); err != nil {
			logger.Error(err, "failed to update KyvernoPolicy statis in NimbusPolicy")
		}

		logger.Info("Dangling KyvernoPolicy deleted", "KyvernoPolicy.Name", kp.Name, "KyvernoPolicy.Namespace", kp.Namespace)
	}
}

func deleteKcp(ctx context.Context, cnpName string) {
	logger := log.FromContext(ctx)
	var kcps kyvernov1.ClusterPolicyList

	if err := k8sClient.List(ctx, &kcps); err != nil {
		logger.Error(err, "failed to list KyvernoClusterPolicies")
		return
	}

	// Kubernetes GC automatically deletes the child when the parent/owner is
	// deleted. So, we don't need to do anything in this case since NimbusPolicy is
	// the owner and when it gets deleted corresponding kps will be automatically
	// deleted.
	for _, kcp := range kcps.Items {
		logger.Info("KyvernoClusterPolicy already deleted due to ClusterNimbusPolicy deletion",
			"KyvernoClusterPolicy.Name", kcp.Name,
			"ClusterNimbusPolicy.Name", cnpName,
		)
	}
}

func deleteDanglingkcps(ctx context.Context, cnp intentv1.ClusterNimbusPolicy, logger logr.Logger) {
	var existingkcps kyvernov1.ClusterPolicyList
	if err := k8sClient.List(ctx, &existingkcps); err != nil {
		logger.Error(err, "failed to list KyvernoClusterPolicies for cleanup")
		return
	}

	var kcpsOwnedByCnp []kyvernov1.ClusterPolicy
	for _, kcp := range existingkcps.Items {
		for _, ownerRef := range kcp.OwnerReferences {
			if ownerRef.Name == cnp.Name && ownerRef.UID == cnp.UID {
				kcpsOwnedByCnp = append(kcpsOwnedByCnp, kcp)
				break
			}
		}
	}
	if len(kcpsOwnedByCnp) == 0 {
		return
	}

	kcpsToDelete := make(map[string]kyvernov1.ClusterPolicy)

	// Populate owned kcps
	for _, kcpOwnedByCnp := range kcpsOwnedByCnp {
		kcpsToDelete[kcpOwnedByCnp.Name] = kcpOwnedByCnp
	}

	for _, nimbusRule := range cnp.Spec.NimbusRules {
		kcpName := cnp.Name + "-" + strings.ToLower(nimbusRule.ID)
		delete(kcpsToDelete, kcpName)
	}

	for kcpName := range kcpsToDelete {
		kcp := kcpsToDelete[kcpName]
		if err := k8sClient.Delete(ctx, &kcp); err != nil {
			logger.Error(err, "failed to delete dangling KyvernoClusterPolicy", "KyvernoClusterPolicy.Name", kcp.Name)
			continue
		}

		logger.Info("Dangling KyvernoClusterPolicy deleted", "KyvernoClusterPolicy.Name", kcp.Name)

		if err := adapterutil.UpdateCnpStatus(ctx, k8sClient, "KyvernoClusterPolicy/"+kcp.Name, cnp.Name, true); err != nil {
			logger.Error(err, "failed to update KyvernoClusterPolicy statis in ClusterNimbusPolicy")
		}

	}
}
