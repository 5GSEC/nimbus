// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"bytes"
	"context"
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/log"

	adapterutil "github.com/5GSEC/nimbus/pkg/adapter/util"
)

func setupClusterNpInformer() cache.SharedIndexInformer {
	clusterNpGvr := schema.GroupVersionResource{
		Group:    "intent.security.nimbus.com",
		Version:  "v1",
		Resource: "clusternimbuspolicies",
	}
	clusterNpInformer := factory.ForResource(clusterNpGvr).Informer()
	return clusterNpInformer
}

// WatchClusterNimbusPolicies watches for create, update and delete events for
// ClusterNimbusPolicies owned by ClusterSecurityIntentBinding and put their info
// on respective channels.
func WatchClusterNimbusPolicies(ctx context.Context, clusterNpChan chan string, deletedClusterNpChan chan string) {
	clusterNpInformer := setupClusterNpInformer()
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

			oldSpec, errOld := oldU.Object["spec"].(map[string]interface{})
			newSpec, errNew := newU.Object["spec"].(map[string]interface{})

			if errOld && errNew {
				oldSpecBytes, _ := json.Marshal(oldSpec)
				newSpecBytes, _ := json.Marshal(newSpec)
				if bytes.Equal(oldSpecBytes, newSpecBytes) {
					return
				}
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
			deletedClusterNpChan <- u.GetName()
		},
	}
	_, err := clusterNpInformer.AddEventHandler(handlers)
	if err != nil {
		logger.Error(err, "failed to add event handlers")
		return
	}
	logger.Info("ClusterNimbusPolicy watcher started")
	clusterNpInformer.Run(ctx.Done())
}
