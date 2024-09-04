// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"context"
	"strings"
	"time"

	"github.com/5GSEC/nimbus/pkg/adapter/common"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-kyverno/utils"

	adapterutil "github.com/5GSEC/nimbus/pkg/adapter/util"
	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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

func kpInformer() cache.SharedIndexInformer {
	kpGvr := schema.GroupVersionResource{
		Group:    "kyverno.io",
		Version:  "v1",
		Resource: "policies",
	}
	informer := factory.ForResource(kpGvr).Informer()
	return informer
}

// WatchKps watches update and delete event for KyvernoPolicies owned by
// NimbusPolicy or ClusterNimbusPolicy and put their info on respective channels.
func WatchKps(ctx context.Context, updatedKpCh, deletedKpCh chan common.Request) {
	logger := log.FromContext(ctx)
	informer := kpInformer()
	handlers := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldU := oldObj.(*unstructured.Unstructured)
			newU := newObj.(*unstructured.Unstructured)
			if adapterutil.IsOrphan(newU.GetOwnerReferences(), "NimbusPolicy") {
				logger.V(4).Info("Ignoring orphan KyvernoPolicy", "KyvernoPolicy.Name", oldU.GetName(), "KyvernoPolicy.Namespace", oldU.GetNamespace(), "Operation", "Update")
				return
			}

			var oldKp kyvernov1.Policy
			var newKp kyvernov1.Policy
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(oldU.Object, &oldKp)
			if err != nil {
				logger.Error(err, "failed to convert to kyverno policy")
				return
			}

			err = runtime.DefaultUnstructuredConverter.FromUnstructured(newU.Object, &newKp)
			if err != nil {
				logger.Error(err, "failed to convert to kyverno policy")
				return
			}

			oldConditions := oldKp.Status.Conditions
			newConditions := newKp.Status.Conditions

			if !strings.Contains(newKp.GetName(), "mutateexisting") {
				if oldU.GetGeneration() == newU.GetGeneration() {
					return
				}
				kpNamespacedName := common.Request{
					Name:      newU.GetName(),
					Namespace: newU.GetNamespace(),
				}
				updatedKpCh <- kpNamespacedName
				return
			}

			// for mutate existing policy
			if oldU.GetGeneration() == newU.GetGeneration() {
				if utils.CheckIfReady(newConditions) && !utils.CheckIfReady(oldConditions) {
					kpNamespacedName := common.Request{
						Name:      newU.GetName(),
						Namespace: newU.GetNamespace(),
					}
					updatedKpCh <- kpNamespacedName
					return
				}
				return
			} else {
				kpNamespacedName := common.Request{
					Name:      newU.GetName(),
					Namespace: newU.GetNamespace(),
				}
				updatedKpCh <- kpNamespacedName
			}
		},
		DeleteFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			if adapterutil.IsOrphan(u.GetOwnerReferences(), "NimbusPolicy") {
				logger.V(4).Info("Ignoring orphan KyvernoPolicy", "KyvernoPolicy.Name", u.GetName(), "KyvernoPolicy.Namespace", u.GetNamespace(), "Operation", "Delete")
				return
			}
			kpNamespacedName := common.Request{
				Name:      u.GetName(),
				Namespace: u.GetNamespace(),
			}
			deletedKpCh <- kpNamespacedName
		},
	}
	_, err := informer.AddEventHandler(handlers)
	if err != nil {
		logger.Error(err, "failed to add event handlers")
		return
	}
	logger.Info("KyvernoPolicy watcher started")
	informer.Run(ctx.Done())
}
