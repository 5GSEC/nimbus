// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package manager

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-kubearmor/processor"
)

var (
	scheme    = runtime.NewScheme()
	np        v1.NimbusPolicy
	k8sClient client.Client
	err       error
)

func init() {
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(kubearmorv1.AddToScheme(scheme))

	k8sClient, err = k8s.New(scheme)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ManageKsps(ctx context.Context, nimbusPolicyCh chan [2]string, nimbusPolicyToDeleteCh chan [2]string, nimbusPolicyUpdateCh chan [2]string, clusterNpChan chan string, clusterNpToDeleteChan chan string) {
	for {
		select {
		case _ = <-ctx.Done():
			close(nimbusPolicyCh)
			close(nimbusPolicyToDeleteCh)
			close(clusterNpChan)
			close(clusterNpToDeleteChan)
			return
		case npToCreate := <-nimbusPolicyCh:
			createKsp(ctx, npToCreate[0], npToCreate[1])
		case npToUpdate := <-nimbusPolicyUpdateCh:
			updateKsp(ctx, npToUpdate[0], npToUpdate[1])
		case npToDelete := <-nimbusPolicyToDeleteCh:
			deleteKsp(ctx, npToDelete[0], npToDelete[1])
		case _ = <-clusterNpChan: // Fixme: CreateKSP based on ClusterNP
			fmt.Println("No-op for ClusterNimbusPolicy")
		case _ = <-clusterNpToDeleteChan: // Fixme: DeleteKSP based on ClusterNP
			fmt.Println("No-op for ClusterNimbusPolicy")
		}
	}
}

func deleteKsp(ctx context.Context, npName, npNamespace string) {
	logger := log.FromContext(ctx)
	ksps := processor.BuildKspsFrom(logger, &np)
	for idx := range ksps {
		ksp := ksps[idx]
		var existingKsp kubearmorv1.KubeArmorPolicy
		err := k8sClient.Get(ctx, types.NamespacedName{Name: ksp.Name, Namespace: ksp.Namespace}, &existingKsp)
		if err != nil {
			if errors.IsNotFound(err) {
				logger.Info("KubeArmorPolicy already deleted, no action needed", "KubeArmorPolicy.Name", ksp.Name, "KubeArmorPolicy.Namespace", ksp.Namespace)
			} else {
				logger.Error(err, "failed to get existing KubeArmorPolicy", "KubeArmorPolicy.Name", ksp.Name, "KubeArmorPolicy.Namespace", ksp.Namespace)
				continue
			}
		} else {
			if err = k8sClient.Delete(ctx, &existingKsp); err != nil {
				logger.Error(err, "failed to delete KubeArmorPolicy", "KubeArmorPolicy.Name", ksp.Name, "KubeArmorPolicy.Namespace", ksp.Namespace)
				return
			}
			logger.Info("KubeArmorPolicy deleted due to NimbusPolicy deletion",
				"KubeArmorPolicy.Name", ksp.Name, "KubeArmorPolicy.Namespace", ksp.Namespace,
				"NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", npNamespace,
			)
		}
	}
}

func createKsp(ctx context.Context, npName, npNamespace string) {
	logger := log.FromContext(ctx)
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: npName, Namespace: npNamespace}, &np); err != nil {
		logger.Error(err, "Failed to get NimbusPolicy", "NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", npNamespace)
		return
	}

	// Check if all strict mode intents are implemented by the adapter.
	allStrictIntentsImplemented := true
	for _, rule := range np.Spec.NimbusRules {
		if rule.Rule.Mode == "strict" && !idpool.IsIdSupportedBy(rule.ID, "kubearmor") {
			allStrictIntentsImplemented = false
			logger.Info("The adapter does not support the strict mode intent", "ID", rule.ID)
			break
		}
	}

	// If there is any unimplemented strict mode intent, skip processing the NimbusPolicy.
	if !allStrictIntentsImplemented {
		logger.Info("Skipping NimbusPolicy processing.", "NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", npNamespace)
		return
	}

	ksps := processor.BuildKspsFrom(logger, &np)
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
		if errors.IsNotFound(err) {
			createOrUpdateKsp(ctx, &ksp, nil, logger)
		} else {
			createOrUpdateKsp(ctx, &ksp, &existingKsp, logger)
		}
	}
}

func updateKsp(ctx context.Context, npName, npNamespace string) {
	logger := log.FromContext(ctx)

	if err := k8sClient.Get(ctx, types.NamespacedName{Name: npName, Namespace: npNamespace}, &np); err != nil {
		logger.Error(err, "Failed to get NimbusPolicy for update", "NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", npNamespace)
		return
	}

	kspsToCreate := processor.BuildKspsFrom(logger, &np)

	for _, kspToCreate := range kspsToCreate {
		var existingKsp kubearmorv1.KubeArmorPolicy
		err := k8sClient.Get(ctx, types.NamespacedName{Name: kspToCreate.Name, Namespace: kspToCreate.Namespace}, &existingKsp)
		if err != nil && errors.IsNotFound(err) {
			createOrUpdateKsp(ctx, &kspToCreate, nil, logger)
		} else if err == nil {
			createOrUpdateKsp(ctx, &kspToCreate, &existingKsp, logger)
		} else {
			logger.Error(err, "failed to get KubeArmorPolicy for update", "KubeArmorPolicy.Name", kspToCreate.Name, "KubeArmorPolicy.Namespace", kspToCreate.Namespace)
		}
	}

	deleteUnnecessaryKsps(ctx, kspsToCreate, npNamespace, logger)
}

func deleteUnnecessaryKsps(ctx context.Context, currentKsps []kubearmorv1.KubeArmorPolicy, namespace string, logger logr.Logger) {
	var existingKsps kubearmorv1.KubeArmorPolicyList
	if err := k8sClient.List(ctx, &existingKsps, client.InNamespace(namespace)); err != nil {
		logger.Error(err, "Failed to list KubeArmorPolicies for cleanup")
		return
	}

	currentKspNames := make(map[string]bool)
	for _, ksp := range currentKsps {
		currentKspNames[ksp.Name] = true
	}

	for _, existingKsp := range existingKsps.Items {
		if _, needed := currentKspNames[existingKsp.Name]; !needed {
			if err := k8sClient.Delete(ctx, &existingKsp); err != nil {
				logger.Error(err, "Failed to delete unnecessary KubeArmorPolicy", "KubeArmorPolicy.Name", existingKsp.Name)
			} else {
				logger.Info("Deleted unnecessary KubeArmorPolicy", "KubeArmorPolicy.Name", existingKsp.Name)
			}
		}
	}
}

func createOrUpdateKsp(ctx context.Context, kspToCreate *kubearmorv1.KubeArmorPolicy, existingKsp *kubearmorv1.KubeArmorPolicy, logger logr.Logger) {
	if existingKsp == nil {
		if err = k8sClient.Create(ctx, kspToCreate); err != nil {
			logger.Error(err, "failed to create KubeArmorPolicy", "KubeArmorPolicy.Name", kspToCreate.Name, "KubeArmorPolicy.Namespace", kspToCreate.Namespace)
			return
		}
		logger.Info("KubeArmorPolicy Created", "KubeArmorPolicy.Name", kspToCreate.Name, "KubeArmorPolicy.Namespace", kspToCreate.Namespace)
	} else {
		kspToCreate.ObjectMeta.ResourceVersion = existingKsp.ObjectMeta.ResourceVersion
		if err = k8sClient.Update(ctx, kspToCreate); err != nil {
			logger.Error(err, "failed to configure existing KubeArmorPolicy", "KubeArmorPolicy.Name", existingKsp.Name, "KubeArmorPolicy.Namespace", existingKsp.Namespace)
			return
		}
		logger.Info("KubeArmorPolicy Configured", "KubeArmorPolicy.Name", existingKsp.Name, "KubeArmorPolicy.Namespace", existingKsp.Namespace)
	}
}
