// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package manager

import (
	"context"
	"fmt"

	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-netpol/processor"
)

var (
	scheme    = runtime.NewScheme()
	np        v1.NimbusPolicy
	k8sClient client.Client
	err       error
)

func init() {
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(netv1.AddToScheme(scheme))

	k8sClient, err = k8s.New(scheme)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ManageNetPols(ctx context.Context, nimbusPolicyCh chan [2]string, nimbusPolicyToDeleteCh chan [2]string, clusterNpChan chan string, clusterNpToDeleteChan chan string) {
	for {
		select {
		case _ = <-ctx.Done():
			close(nimbusPolicyCh)
			close(nimbusPolicyToDeleteCh)
			close(clusterNpChan)
			close(clusterNpToDeleteChan)
			return
		case createdNp := <-nimbusPolicyCh:
			createNetworkPolicy(ctx, createdNp[0], createdNp[1])
		case deletedNp := <-nimbusPolicyToDeleteCh:
			deleteNetworkPolicy(ctx, deletedNp[0], deletedNp[1])
		case _ = <-clusterNpChan: // Fixme: Create netpol based on ClusterNP
			fmt.Println("No-op for ClusterNimbusPolicy")
		case _ = <-clusterNpToDeleteChan: // Fixme: Delete netpol based on ClusterNP
			fmt.Println("No-op for ClusterNimbusPolicy")
		}
	}
}

func createNetworkPolicy(ctx context.Context, npName, npNamespace string) {
	logger := log.FromContext(ctx)
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: npName, Namespace: npNamespace}, &np); err != nil {
		logger.Error(err, "failed to get NimbusPolicy", "NimbusPolicy.Name", npName[0], "NimbusPolicy.Namespace", npName[1])
		return
	}

	netPols := processor.BuildNetPolsFrom(logger, np)
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
		if errors.IsNotFound(err) {
			if err = k8sClient.Create(ctx, &netpol); err != nil {
				logger.Error(err, "failed to create NetworkPolicy", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
				return
			}
			logger.Info("NetworkPolicy created", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
		} else {
			netpol.ObjectMeta.ResourceVersion = existingNetpol.ObjectMeta.ResourceVersion
			if err = k8sClient.Update(ctx, &netpol); err != nil {
				logger.Error(err, "failed to configure existing NetworkPolicy", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
				return
			}
			logger.Info("NetworkPolicy configured", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
		}
	}
}

func deleteNetworkPolicy(ctx context.Context, npName, npNamespace string) {
	logger := log.FromContext(ctx)
	netPols := processor.BuildNetPolsFrom(logger, np)
	for idx := range netPols {
		netpol := netPols[idx]
		var existingNetpol netv1.NetworkPolicy
		err := k8sClient.Get(ctx, types.NamespacedName{Name: npName, Namespace: npNamespace}, &existingNetpol)
		if err != nil {
			if errors.IsNotFound(err) {
				logger.Info("NetworkPolicy already deleted, no action needed", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
			} else {
				logger.Error(err, "failed to get existing NetworkPolicy", "NetworkPolicy.Name", netpol.Name, "NetworkPolicy.Namespace", netpol.Namespace)
				continue
			}
		}
	}
}
