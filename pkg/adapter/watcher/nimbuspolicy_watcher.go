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

// WatchNimbusPolicies watches for create, update and delete events for
// NimbusPolicies owned by SecurityIntentBinding and put their info on respective
// channels.
// ownerKind indicates which owners of the NimbusPolicy are fine
func WatchNimbusPolicies(ctx context.Context, npCh, deleteNpCh chan common.Request, ownerKind ...string) {
	nimbusPolicyInformer := npInformer()
	logger := log.FromContext(ctx)

	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			if adapterutil.IsOrphan(u.GetOwnerReferences(), ownerKind...) {
				logger.V(4).Info("Ignoring orphan NimbusPolicy", "NimbusPolicy.Name", u.GetName(), "NimbusPolicy.Namespace", u.GetNamespace(), "Operation", "Create")
				return
			}
			npNamespacedName := common.Request{
				Name:      u.GetName(),
				Namespace: u.GetNamespace(),
			}
			logger.Info("NimbusPolicy found", "NimbusPolicy.Name", u.GetName(), "NimbusPolicy.Namespace", u.GetNamespace())
			npCh <- npNamespacedName
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldU := oldObj.(*unstructured.Unstructured)
			newU := newObj.(*unstructured.Unstructured)

			if adapterutil.IsOrphan(newU.GetOwnerReferences(), ownerKind...) {
				logger.V(4).Info("Ignoring orphan NimbusPolicy", "NimbusPolicy.Name", oldU.GetName(), "NimbusPolicy.Namespace", oldU.GetNamespace(), "Operation", "Update")
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
			npNamespacedName := common.Request{
				Name:      newU.GetName(),
				Namespace: newU.GetNamespace(),
			}
			logger.Info("NimbusPolicy modified", "NimbusPolicy.Name", newU.GetName(), "NimbusPolicy.Namespace", newU.GetNamespace())
			npCh <- npNamespacedName
		},
		DeleteFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			if adapterutil.IsOrphan(u.GetOwnerReferences(), ownerKind...) {
				logger.V(4).Info("Ignoring orphan NimbusPolicy", "NimbusPolicy.Name", u.GetName(), "NimbusPolicy.Namespace", u.GetNamespace(), "Operation", "Delete")
				return
			}
			npNamespacedName := common.Request{
				Name:      u.GetName(),
				Namespace: u.GetNamespace(),
			}
			logger.Info("NimbusPolicy deleted", "NimbusPolicy.Name", u.GetName(), "NimbusPolicy.Namespace", u.GetNamespace())
			deleteNpCh <- npNamespacedName
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
