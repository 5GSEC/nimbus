// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"context"
	"time"

	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/log"

	adapterutil "github.com/5GSEC/nimbus/pkg/adapter/util"
)

func init() {
	factory = dynamicinformer.NewDynamicSharedInformerFactory(k8s.NewDynamicClient(), time.Minute)
}

func kcpInformer() cache.SharedIndexInformer {
	kcpGvr := schema.GroupVersionResource{
		Group:    "kyverno.io",
		Version:  "v1",
		Resource: "clusterpolicies",
	}
	informer := factory.ForResource(kcpGvr).Informer()
	return informer
}

// WatchKcps watches update and delete event for KyvernoClusterPolicies owned by
// NimbusPolicy or ClusterNimbusPolicy and put their info on respective channels.
func WatchKcps(ctx context.Context, updatedKcpCh, deletedKcpCh chan string) {
	logger := log.FromContext(ctx)
	informer := kcpInformer()
	handlers := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldU := oldObj.(*unstructured.Unstructured)
			newU := newObj.(*unstructured.Unstructured)

			if adapterutil.IsOrphan(newU.GetOwnerReferences(), "ClusterNimbusPolicy") {
				logger.V(4).Info("Ignoring orphan KyvernoClusterPolicy", "KyvernoClusterPolicy.Name", oldU.GetName(), "Operation", "Update")
				return
			}

			if oldU.GetGeneration() == newU.GetGeneration() {
				return
			}

			kcpName := newU.GetName()
			updatedKcpCh <- kcpName
		},
		DeleteFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			if adapterutil.IsOrphan(u.GetOwnerReferences(), "ClusterNimbusPolicy") {
				logger.V(4).Info("Ignoring orphan KyvernoClusterPolicy", "KyvernoClusterPolicy.Name", u.GetName(), "Operation", "Delete")
				return
			}
			kcpName := u.GetName()
			deletedKcpCh <- kcpName
		},
	}
	_, err := informer.AddEventHandler(handlers)
	if err != nil {
		logger.Error(err, "failed to add event handlers")
		return
	}
	logger.Info("KyvernoClusterPolicy watcher started")
	informer.Run(ctx.Done())
}
