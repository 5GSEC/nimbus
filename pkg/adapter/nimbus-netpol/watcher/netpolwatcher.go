// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/5GSEC/nimbus/pkg/adapter/common"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	adapterutil "github.com/5GSEC/nimbus/pkg/adapter/util"
)

var (
	factory dynamicinformer.DynamicSharedInformerFactory
)

func init() {
	factory = dynamicinformer.NewDynamicSharedInformerFactory(k8s.NewDynamicClient(), time.Minute)
}

func netpolInformer() cache.SharedIndexInformer {
	kspGvr := schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "networkpolicies",
	}
	informer := factory.ForResource(kspGvr).Informer()
	return informer
}

// WatchNetpols watches for update and delete events for NetworkPolicies owned by
// NimbusPolicy or ClusterNimbusPolicy and put their info on respective channels.
func WatchNetpols(ctx context.Context, updatedNetpolCh, deletedNetpolCh chan common.Request) {
	logger := log.FromContext(ctx)
	informer := netpolInformer()
	handlers := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldU := oldObj.(*unstructured.Unstructured)
			newU := newObj.(*unstructured.Unstructured)

			if adapterutil.IsOrphan(newU.GetOwnerReferences(), "NimbusPolicy") {
				logger.V(4).Info("Ignoring orphan NetworkPolicy", "NetworkPolicy.Name", oldU.GetName(), "NetworkPolicy.Namespace", oldU.GetNamespace(), "Operation", "Update")
				return
			}

			if oldU.GetGeneration() == newU.GetGeneration() {
				return
			}

			netpolNamespacedName := common.Request{
				Name:      newU.GetName(),
				Namespace: newU.GetNamespace(),
			}
			updatedNetpolCh <- netpolNamespacedName
		},
		DeleteFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			if adapterutil.IsOrphan(u.GetOwnerReferences(), "NimbusPolicy") {
				logger.V(4).Info("Ignoring orphan NetworkPolicy", "NetworkPolicy.Name", u.GetName(), "NetworkPolicy.Namespace", u.GetNamespace(), "Operation", "Delete")
				return
			}
			netpolNamespacedName := common.Request{
				Name:      u.GetName(),
				Namespace: u.GetNamespace(),
			}
			deletedNetpolCh <- netpolNamespacedName
		},
	}
	_, err := informer.AddEventHandler(handlers)
	if err != nil {
		logger.Error(err, "failed to add event handlers")
		return
	}
	logger.Info("NetworkPolicy watcher started")
	informer.Run(ctx.Done())
}
