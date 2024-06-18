// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"

	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
)

var factory dynamicinformer.DynamicSharedInformerFactory

func init() {
	k8sClient := k8s.NewDynamicClient()
	factory = dynamicinformer.NewDynamicSharedInformerFactory(k8sClient, time.Minute)
}

func npInformer() cache.SharedIndexInformer {
	nimbusPolicyGvr := schema.GroupVersionResource{
		Group:    "intent.security.nimbus.com",
		Version:  "v1alpha1",
		Resource: "nimbuspolicies",
	}
	nimbusPolicyInformer := factory.ForResource(nimbusPolicyGvr).Informer()
	return nimbusPolicyInformer
}

func clusterNpInformer() cache.SharedIndexInformer {
	clusterNpGvr := schema.GroupVersionResource{
		Group:    "intent.security.nimbus.com",
		Version:  "v1alpha1",
		Resource: "clusternimbuspolicies",
	}
	clusterNimbusPolicyInformer := factory.ForResource(clusterNpGvr).Informer()
	return clusterNimbusPolicyInformer
}
