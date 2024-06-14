// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/5GSEC/nimbus/pkg/adapter/common"
	adapterutil "github.com/5GSEC/nimbus/pkg/adapter/util"
)

// WatchNimbusPolicies watches for create, update and delete events for
// NimbusPolicies owned by SecurityIntentBinding and put their info on respective
// channels.
// ownerKind indicates which owners of the NimbusPolicy are fine
func WatchNimbusPolicies(ctx context.Context, npCh chan common.Request, deleteNpCh chan *unstructured.Unstructured, ownerKind ...string) {
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

			if oldU.GetGeneration() == newU.GetGeneration() {
				return
			}

			npNamespacedName := common.Request{
				Name:      newU.GetName(),
				Namespace: newU.GetNamespace(),
			}
			logger.Info("NimbusPolicy modified", "NimbusPolicy.Name", newU.GetName(), "NimbusPolicy.Namespace", newU.GetNamespace())
			npCh <- npNamespacedName
		},
		DeleteFunc: func(obj interface{}) {
			deletedObj := obj.(*unstructured.Unstructured)
			if adapterutil.IsOrphan(deletedObj.GetOwnerReferences(), ownerKind...) {
				logger.V(4).Info("Ignoring orphan NimbusPolicy", "NimbusPolicy.Name", deletedObj.GetName(), "NimbusPolicy.Namespace", deletedObj.GetNamespace(), "Operation", "Delete")
				return
			}
			logger.Info("NimbusPolicy deleted", "NimbusPolicy.Name", deletedObj.GetName(), "NimbusPolicy.Namespace", deletedObj.GetNamespace())
			deleteNpCh <- deletedObj
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
