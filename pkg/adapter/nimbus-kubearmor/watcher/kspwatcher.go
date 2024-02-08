// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"bytes"
	"context"
	"encoding/json"
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

func kspInformer() cache.SharedIndexInformer {
	kspGvr := schema.GroupVersionResource{
		Group:    "security.kubearmor.com",
		Version:  "v1",
		Resource: "kubearmorpolicies",
	}
	informer := factory.ForResource(kspGvr).Informer()
	return informer
}

// WatchKsps watches update and delete event for KubeArmorPolicies owned by
// NimbusPolicy or ClusterNimbusPolicy and put their info on respective channels.
func WatchKsps(ctx context.Context, updatedKspCh, deletedKspCh chan common.Request) {
	logger := log.FromContext(ctx)
	informer := kspInformer()
	handlers := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldU := oldObj.(*unstructured.Unstructured)
			newU := newObj.(*unstructured.Unstructured)

			if adapterutil.IsOrphan(newU.GetOwnerReferences(), "NimbusPolicy") {
				logger.V(4).Info("Ignoring orphan KubeArmorPolicy", "KubeArmorPolicy.Name", oldU.GetName(), "KubeArmorPolicy.Namespace", oldU.GetNamespace(), "Operation", "Update")
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

			kspNamespacedName := common.Request{
				Name:      newU.GetName(),
				Namespace: newU.GetNamespace(),
			}
			updatedKspCh <- kspNamespacedName
		},
		DeleteFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			if adapterutil.IsOrphan(u.GetOwnerReferences(), "NimbusPolicy") {
				logger.V(4).Info("Ignoring orphan KubeArmorPolicy", "KubeArmorPolicy.Name", u.GetName(), "KubeArmorPolicy.Namespace", u.GetNamespace(), "Operation", "Delete")
				return
			}
			kspNamespacedName := common.Request{
				Name:      u.GetName(),
				Namespace: u.GetNamespace(),
			}
			deletedKspCh <- kspNamespacedName
		},
	}
	_, err := informer.AddEventHandler(handlers)
	if err != nil {
		logger.Error(err, "failed to add event handlers")
		return
	}
	logger.Info("KubeArmorPolicy watcher started")
	informer.Run(ctx.Done())
}
