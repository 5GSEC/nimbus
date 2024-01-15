// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func WatchClusterNimbusPolicies(ctx context.Context, clusterNpChan chan string, clusterNpToDeleteChan chan string) {
	logger := log.FromContext(ctx)
	clusterNpInformer := setupClusterNpInformer()
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			logger.Info("ClusterNimbusPolicy found", "ClusterNimbusPolicy.Name", u.GetName())
			clusterNpChan <- u.GetName()
		},
		DeleteFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			logger.Info("ClusterNimbusPolicy deleted", "ClusterNimbusPolicy.Name", u.GetName())
			clusterNpToDeleteChan <- u.GetName()
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

func setupClusterNpInformer() cache.SharedIndexInformer {
	clusterNpGvr := schema.GroupVersionResource{
		Group:    "intent.security.nimbus.com",
		Version:  "v1",
		Resource: "clusternimbuspolicies",
	}
	clusterNpInformer := factory.ForResource(clusterNpGvr).Informer()
	return clusterNpInformer
}
