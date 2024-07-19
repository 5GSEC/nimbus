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
)

var (
	factory dynamicinformer.DynamicSharedInformerFactory
)

func init() {
	factory = dynamicinformer.NewDynamicSharedInformerFactory(k8s.NewDynamicClient(), time.Minute)
}

func dspInformer() cache.SharedIndexInformer {
	dspGvr := schema.GroupVersionResource{
		Group:    "security.accuknox.com",
		Version:  "v1",
		Resource: "discoveredpolicies",
	}
	return factory.ForResource(dspGvr).Informer()
}

// WatchDsps watches for create and update event for DiscoveredPolicies and put
// their info on respective channels.
func WatchDsps(ctx context.Context, createdUpdatedDsps chan common.Request) {
	logger := log.FromContext(ctx)
	informer := dspInformer()

	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			dsp := obj.(*unstructured.Unstructured)
			createdUpdatedDsps <- common.Request{
				Name:      dsp.GetName(),
				Namespace: dsp.GetNamespace(),
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldDsp := oldObj.(*unstructured.Unstructured)
			newDsp := newObj.(*unstructured.Unstructured)
			if oldDsp.GetGeneration() != newDsp.GetGeneration() {
				createdUpdatedDsps <- common.Request{
					Name:      newDsp.GetName(),
					Namespace: newDsp.GetNamespace(),
				}
			}
		},
	}

	if _, err := informer.AddEventHandler(handlers); err != nil {
		logger.Error(err, "failed to add dsp event handler")
		return
	}

	logger.V(2).Info("Discovered Policy watcher started")
	informer.Run(ctx.Done())
}
