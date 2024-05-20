// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"context"
	"time"

	"github.com/5GSEC/nimbus/pkg/adapter/common"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	factory dynamicinformer.DynamicSharedInformerFactory
)

func init() {
	factory = dynamicinformer.NewDynamicSharedInformerFactory(k8s.NewDynamicClient(), time.Minute)
}

func podInformer() cache.SharedIndexInformer {
	podGvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	informer := factory.ForResource(podGvr).Informer()
	return informer
}

// WatchPods watches for create and update events for Pods that match the selector of active Nimbus Policies
func WatchPods(ctx context.Context, updatedPodCh chan common.Request) {
	informer := podInformer()
	logger := log.FromContext(ctx)

	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			podReq := common.Request{
				Name:      u.GetName(),
				Namespace: u.GetNamespace(),
			}
			updatedPodCh <- podReq // Send pod information to the channel
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newU := newObj.(*unstructured.Unstructured)
			oldU := oldObj.(*unstructured.Unstructured)
			if newU.GetResourceVersion() == oldU.GetResourceVersion() {
				return // No actual update in terms of data
			}
			podReq := common.Request{
				Name:      newU.GetName(),
				Namespace: newU.GetNamespace(),
			}
			updatedPodCh <- podReq // Send updated pod information to the channel
		},
		/*
			DeleteFunc: func(obj interface{}) {
				u := obj.(*unstructured.Unstructured)
				podReq := common.Request{
					Name:      u.GetName(),
					Namespace: u.GetNamespace(),
				}
				deletedPodCh <- podReq // Send pod deletion information to the channel
			},
		*/
	}
	_, err := informer.AddEventHandler(handlers)
	if err != nil {
		logger.Error(err, "failed to add event handlers")
		return
	}
	logger.Info("Pod watcher started")
	informer.Run(ctx.Done())
}
