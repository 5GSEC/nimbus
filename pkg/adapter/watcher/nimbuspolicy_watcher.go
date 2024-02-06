// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"context"
	"encoding/json"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
)

var factory dynamicinformer.DynamicSharedInformerFactory

func init() {
	k8sClient := k8s.NewDynamicClient()
	factory = dynamicinformer.NewDynamicSharedInformerFactory(k8sClient, time.Minute)
}

func WatchNimbusPolicies(ctx context.Context, nimbusPolicyCh chan [2]string, nimbusPolicyToDeleteCh chan [2]string, nimbusPolicyUpdateCh chan [2]string) {
	nimbusPolicyInformer := setupNpInformer()
	logger := log.FromContext(ctx)
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			npNamespacedName := [2]string{u.GetName(), u.GetNamespace()}
			nimbusPolicyCh <- npNamespacedName
			logger.Info("NimbusPolicy found", "NimbusPolicy.Name", npNamespacedName[0], "NimbusPolicy.Namespace", npNamespacedName[1])
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldU := oldObj.(*unstructured.Unstructured)
			newU := newObj.(*unstructured.Unstructured)

			// spec을 JSON으로 마샬링하여 문자열 비교
			oldSpec, errOld := oldU.Object["spec"].(map[string]interface{})
			newSpec, errNew := newU.Object["spec"].(map[string]interface{})

			if errOld && errNew {
				oldSpecBytes, _ := json.Marshal(oldSpec)
				newSpecBytes, _ := json.Marshal(newSpec)

				if string(oldSpecBytes) == string(newSpecBytes) {
					return
				}
			}

			npNamespacedName := [2]string{newU.GetName(), newU.GetNamespace()}
			nimbusPolicyUpdateCh <- npNamespacedName
			logger.Info("NimbusPolicy update detected, event sent to channel", "NimbusPolicy.Name", npNamespacedName[0], "NimbusPolicy.Namespace", npNamespacedName[1])
		},
		DeleteFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			npNamespacedName := [2]string{u.GetName(), u.GetNamespace()}
			nimbusPolicyToDeleteCh <- npNamespacedName
			logger.Info("NimbusPolicy deleted", "NimbusPolicy.Name", npNamespacedName[0], "NimbusPolicy.Namespace", npNamespacedName[1])
		},
	}
	_, err := nimbusPolicyInformer.AddEventHandler(handlers)
	if err != nil {
		logger.Error(err, "failed to add event handlers")
		return
	}
	logger.Info("NimbusPolicy watcher started")
	nimbusPolicyInformer.Run(ctx.Done())
}

func setupNpInformer() cache.SharedIndexInformer {
	nimbusPolicyGvr := schema.GroupVersionResource{
		Group:    "intent.security.nimbus.com",
		Version:  "v1",
		Resource: "nimbuspolicies",
	}
	nimbusPolicyInformer := factory.ForResource(nimbusPolicyGvr).Informer()
	return nimbusPolicyInformer
}
