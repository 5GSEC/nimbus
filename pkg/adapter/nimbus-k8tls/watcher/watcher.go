// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package watcher

import (
	"context"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/5GSEC/nimbus/pkg/adapter/common"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	adapterutil "github.com/5GSEC/nimbus/pkg/adapter/util"
)

var (
	factory informers.SharedInformerFactory
)

func init() {
	factory = informers.NewSharedInformerFactory(k8s.NewOrDieStaticClient(), time.Minute)
}

func cronJobInformer() cache.SharedIndexInformer {
	return factory.Batch().V1().CronJobs().Informer()
}

// WatchCronJobs watches for update and delete events for Kubernetes CronJobs
// owned by ClusterNimbusPolicy and put their info on corresponding channels.
func WatchCronJobs(ctx context.Context, updatedCronJobCh, deletedCronJobCh chan common.Request) {
	logger := log.FromContext(ctx)
	informer := cronJobInformer()
	handlers := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldCronJob := oldObj.(*batchv1.CronJob)
			newCronJob := newObj.(*batchv1.CronJob)

			if adapterutil.IsOrphan(newCronJob.GetOwnerReferences(), "ClusterNimbusPolicy") {
				logger.V(4).Info("Ignoring orphan CronJob", "CronJob.Name", oldCronJob.GetName(), "CronJob.Namespace", oldCronJob.GetNamespace(), "Operation", "Update")
				return
			}

			if oldCronJob.GetGeneration() == newCronJob.GetGeneration() {
				return
			}

			cronJobNamespacedName := common.Request{
				Name:      newCronJob.GetName(),
				Namespace: newCronJob.GetNamespace(),
			}
			updatedCronJobCh <- cronJobNamespacedName
		},
		DeleteFunc: func(obj interface{}) {
			cronJob := obj.(*batchv1.CronJob)
			if adapterutil.IsOrphan(cronJob.GetOwnerReferences(), "ClusterNimbusPolicy") {
				logger.V(4).Info("Ignoring orphan CronJob", "CronJob.Name", cronJob.GetName(), "CronJob.Namespace", cronJob.GetNamespace(), "Operation", "Delete")
				return
			}
			cronJobNamespacedName := common.Request{
				Name:      cronJob.GetName(),
				Namespace: cronJob.GetNamespace(),
			}
			deletedCronJobCh <- cronJobNamespacedName
		},
	}
	if _, err := informer.AddEventHandler(handlers); err != nil {
		logger.Error(err, "failed to add event handler")
		return
	}
	logger.Info("Kubernetes CronJob watcher started")
	informer.Run(ctx.Done())
}
