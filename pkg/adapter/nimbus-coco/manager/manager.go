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

	v1alpha1 "github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-coco/processor"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-coco/watcher"
	adapterutil "github.com/5GSEC/nimbus/pkg/adapter/util"
	globalwatcher "github.com/5GSEC/nimbus/pkg/adapter/watcher"
)

var (
	scheme    = runtime.NewScheme()
	k8sClient client.Client
)

func init() {
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(kyvernov1.AddToScheme(scheme))
	k8sClient = k8s.NewOrDie(scheme)
}

func Run(ctx context.Context) {
	clusterNpChan := make(chan string)
	deletedClusterNpChan := make(chan string)
	go globalwatcher.WatchClusterNimbusPolicies(ctx, clusterNpChan, deletedClusterNpChan)

	updatedKcpCh := make(chan string)
	deletedKcpCh := make(chan string)
	go watcher.WatchKcps(ctx, updatedKcpCh, deletedKcpCh)

	for {
		select {
		case <-ctx.Done():
			close(clusterNpChan)
			close(deletedClusterNpChan)
			close(updatedKcpCh)
			close(deletedKcpCh)
			return
		case createdCnp := <-clusterNpChan:
			createOrUpdateKcp(ctx, createdCnp)
		case deletedCnp := <-deletedClusterNpChan:
			deleteKcp(ctx, deletedCnp)
		case updatedKcp := <-updatedKcpCh:
			reconcileKcp(ctx, updatedKcp, false)
		case deletedKcp := <-deletedKcpCh:
			reconcileKcp(ctx, deletedKcp, true)
		}

	}
}
func reconcileKcp(ctx context.Context, kcpName string, deleted bool) {
	logger := log.FromContext(ctx)
	cnpName := adapterutil.ExtractClusterNpName(kcpName)
	var cnp v1alpha1.ClusterNimbusPolicy
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

func createOrUpdateKcp(ctx context.Context, cnpName string) {
	logger := log.FromContext(ctx)
	var cnp v1alpha1.ClusterNimbusPolicy
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

func deleteDanglingkcps(ctx context.Context, cnp v1alpha1.ClusterNimbusPolicy, logger logr.Logger) {
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
