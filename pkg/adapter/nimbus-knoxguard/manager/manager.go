// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package manager

import (
	"context"
	"strings"
	"time"

	v1alpha1 "github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/common"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-kyverno/processor"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-kyverno/watcher"
	adapterutil "github.com/5GSEC/nimbus/pkg/adapter/util"
	globalwatcher "github.com/5GSEC/nimbus/pkg/adapter/watcher"
	"github.com/go-logr/logr"
	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	scheme    = runtime.NewScheme()
	k8sClient client.Client
	metadataClient metadata.Interface
)

func init() {
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(kyvernov1.AddToScheme(scheme))
	utilruntime.Must(kubearmorv1.AddToScheme(scheme))
	utilruntime.Must(netv1.AddToScheme(scheme))
	k8sClient = k8s.NewOrDie(scheme)
	metadataClient = k8s.NewMetadataClient()
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
			return
		case createdNp := <-npCh:
			createOrUpdatePolicies(ctx, createdNp.Name, createdNp.Namespace)
		case createdCnp := <-clusterNpChan:
			createOrUpdateClusterPolicies(ctx, createdCnp)
		case deletedNp := <-deletedNpCh:
			deletePolicies(ctx, deletedNp.Name, deletedNp.Namespace)
		case deletedCnp := <-deletedClusterNpChan:
			deleteClusterPolicies(ctx, deletedCnp)
		case updatedKp := <-updatedKpCh:
			reconcilePolicies(ctx, updatedKp.Name, updatedKp.Namespace, true)
		case updatedKcp := <-updatedKcpCh:
			reconcileClusterPolicies(ctx, updatedKcp, false)
		case deletedKcp := <-deletedKcpCh:
			reconcileClusterPolicies(ctx, deletedKcp, true)
		case deletedKp := <-deletedKpCh:
			reconcilePolicies(ctx, deletedKp.Name, deletedKp.Namespace, true)
		}

	}
}

func reconcilePolicies(ctx context.Context, kpName, namespace string, deleted bool) {
	logger := log.FromContext(ctx)
	npName := adapterutil.ExtractNpName(kpName)
	var np v1alpha1.NimbusPolicy
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
	createOrUpdatePolicies(ctx, npName, namespace)
}

func reconcileClusterPolicies(ctx context.Context, kcpName string, deleted bool) {
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

func createOrUpdatePolicies(ctx context.Context, npName, npNamespace string) {

	logger := log.FromContext(ctx)
	var np v1alpha1.NimbusPolicy
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: npName, Namespace: npNamespace}, &np); err != nil {
		logger.Error(err, "failed to get NimbusPolicy", "NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", npNamespace)
		return
	}

	indexers := cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
	}
	options := func(options *metav1.ListOptions) {
		options.LabelSelector = labels.Set(np.Spec.Selector.MatchLabels).String()
	}

	informer := metadatainformer.NewFilteredMetadataInformer(metadataClient,
		schema.GroupVersionResource{
			Group: "",
			Version: "v1",
			Resource: "Pod",
		},
		metav1.NamespaceAll,
		10*time.Minute,
		indexers,
		options,
	)

	if adapterutil.IsOrphan(np.GetOwnerReferences(), "SecurityIntentBinding") {
		logger.V(4).Info("Ignoring orphan NimbusPolicy", "NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", npNamespace)
		return
	}

	deleteDanglingPolicies(ctx, np, logger)

	// TODO: fetch the specific policies from the images which are being added inside the nimbus and then receive the list of 
	// CVE's along with the patch from knoxguard
	// kps := processor.BuildKpsFrom(logger, &np)
	
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

func deleteKsp(ctx context.Context, npName, npNamespace string) {
	logger := log.FromContext(ctx)
	var ksps kubearmorv1.KubeArmorPolicyList

	if err := k8sClient.List(ctx, &ksps, &client.ListOptions{Namespace: npNamespace}); err != nil {
		logger.Error(err, "failed to list KubeArmorPolicies")
		return
	}

	// Kubernetes GC automatically deletes the child when the parent/owner is
	// deleted. So, we don't need to do anything in this case since NimbusPolicy is
	// the owner and when it gets deleted corresponding KSPs will be automatically
	// deleted.
	for _, ksp := range ksps.Items {
		logger.Info("KubeArmorPolicy already deleted due to NimbusPolicy deletion",
			"KubeArmorPolicy.Name", ksp.Name, "KubeArmorPolicy.Namespace", ksp.Namespace,
			"NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", npNamespace,
		)
	}
}

func deleteNetworkPolicy(ctx context.Context, npName, npNamespace string) {
	logger := log.FromContext(ctx)
	var netpols netv1.NetworkPolicyList

	if err := k8sClient.List(ctx, &netpols, &client.ListOptions{Namespace: npNamespace}); err != nil {
		logger.Error(err, "failed to list NetworkPolicies")
		return
	}

	// Kubernetes GC automatically deletes the child when the parent/owner is
	// deleted. So, we don't need to do anything in this case since NimbusPolicy is
	// the owner and when it gets deleted corresponding NetworkPolicies will be automatically
	// deleted.
	for _, netpol := range netpols.Items {
		logger.Info("NetworkPolicy already deleted due to NimbusPolicy deletion",
			"NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace,
			"NetworkPolicy.Name", npName, "NetworkPolicy.Namespace", npNamespace,
		)
	}
}

func deleteDanglingPolicies(ctx context.Context, np v1alpha1.NimbusPolicy, logger logr.Logger) {
	var existingkps kyvernov1.PolicyList
	if err := k8sClient.List(ctx, &existingkps, client.InNamespace(np.Namespace)); err != nil {
		logger.Error(err, "failed to list KyvernoPolicies for cleanup")
		return
	}

	var existingKsps kubearmorv1.KubeArmorPolicyList
	if err := k8sClient.List(ctx, &existingKsps, client.InNamespace(np.Namespace)); err != nil {
		logger.Error(err, "failed to list KubeArmorPolicies for cleanup")
		return
	}

	var existingNetpols netv1.NetworkPolicyList
	if err := k8sClient.List(ctx, &existingNetpols, client.InNamespace(np.Namespace)); err != nil {
		logger.Error(err, "failed to list NetworkPolicies for cleanup")
		return
	}

	var kpsOwnedByNp []kyvernov1.Policy
	var kspsOwnedByNp []kubearmorv1.KubeArmorPolicy
	var netpolsOwnedByNp []netv1.NetworkPolicy

	for _, kp := range existingkps.Items {
		for _, ownerRef := range kp.OwnerReferences {
			if ownerRef.Name == np.Name && ownerRef.UID == np.UID {
				kpsOwnedByNp = append(kpsOwnedByNp, kp)
				break
			}
		}
	}

	for _, netpol := range existingNetpols.Items {
		for _, ownerRef := range netpol.OwnerReferences {
			if ownerRef.Name == np.Name && ownerRef.UID == np.UID {
				netpolsOwnedByNp = append(netpolsOwnedByNp, netpol)
				break
			}
		}
	}

	for _, ksp := range existingKsps.Items {
		for _, ownerRef := range ksp.OwnerReferences {
			if ownerRef.Name == np.Name && ownerRef.UID == np.UID {
				kspsOwnedByNp = append(kspsOwnedByNp, ksp)
				break
			}
		}
	}

	if len(kpsOwnedByNp) != 0 {
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

	if len(kspsOwnedByNp) != 0 {
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

	if len(netpolsOwnedByNp) != 0 {
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

	return
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
