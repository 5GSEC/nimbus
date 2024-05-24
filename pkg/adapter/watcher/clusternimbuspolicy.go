// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/log"

	adapterutil "github.com/5GSEC/nimbus/pkg/adapter/util"
)

// WatchClusterNimbusPolicies watches for create, update and delete events for
// ClusterNimbusPolicies owned by ClusterSecurityIntentBinding and put their info
// on respective channels.
func WatchClusterNimbusPolicies(ctx context.Context, clusterNpChan chan string, deletedClusterNpChan chan *unstructured.Unstructured) {
	clusterNimbusPolicyInformer := clusterNpInformer()
	logger := log.FromContext(ctx)

	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			if adapterutil.IsOrphan(u.GetOwnerReferences(), "ClusterSecurityIntentBinding") {
				logger.V(4).Info("Ignoring orphan ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", u.GetName(), "Operation", "Create")
				return
			}
			logger.Info("ClusterNimbusPolicy found", "ClusterNimbusPolicy.Name", u.GetName())
			clusterNpChan <- u.GetName()
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldU := oldObj.(*unstructured.Unstructured)
			newU := newObj.(*unstructured.Unstructured)

			if adapterutil.IsOrphan(newU.GetOwnerReferences(), "ClusterSecurityIntentBinding") {
				logger.V(4).Info("Ignoring orphan ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", oldU.GetName(), "Operation", "Update")
				return
			}

			if oldU.GetGeneration() == newU.GetGeneration() {
				return
			}

			logger.Info("ClusterNimbusPolicy modified", "ClusterNimbusPolicy.Name", newU.GetName())
			clusterNpChan <- newU.GetName()
		},
		DeleteFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			if adapterutil.IsOrphan(u.GetOwnerReferences(), "ClusterSecurityIntentBinding") {
				logger.V(4).Info("Ignoring orphan ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", u.GetName(), "Operation", "Delete")
				return
			}
			logger.Info("ClusterNimbusPolicy deleted", "ClusterNimbusPolicy.Name", u.GetName())
			deletedClusterNpChan <- u
		},
	}
	_, err := clusterNimbusPolicyInformer.AddEventHandler(handlers)
	if err != nil {
		logger.Error(err, "failed to add event handlers")
		return
	}
	logger.Info("ClusterNimbusPolicy watcher started")
	clusterNimbusPolicyInformer.Run(ctx.Done())
}
