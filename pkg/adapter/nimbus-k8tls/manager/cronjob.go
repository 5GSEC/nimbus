// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package manager

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/5GSEC/nimbus/api/v1alpha1"
	adapterutil "github.com/5GSEC/nimbus/pkg/adapter/util"
)

func createOrUpdateCj(ctx context.Context, logger logr.Logger, cwnp v1alpha1.ClusterNimbusPolicy, cronJob *batchv1.CronJob) {
	cronJob.Namespace = K8tlsNamespace
	cronJob.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName = k8tls
	if err := ctrl.SetControllerReference(&cwnp, cronJob, scheme); err != nil {
		logger.Error(err, "failed to set OwnerReference on Kubernetes CronJob", "CronJob.Name", cronJob.Name)
		return
	}

	var existingCronJob batchv1.CronJob
	err := k8sClient.Get(ctx, types.NamespacedName{Name: cronJob.Name, Namespace: cronJob.Namespace}, &existingCronJob)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "failed to get Kubernetes CronJob", "CronJob.Name", cronJob.Name, "CronJob.Namespace", cronJob.Namespace)
		return
	}

	if err != nil {
		if errors.IsNotFound(err) {
			if err = k8sClient.Create(ctx, cronJob); err != nil {
				logger.Error(err, "failed to create Kubernetes CronJob", "CronJob.Name", cronJob.Name, "CronJob.Namespace", cronJob.Namespace)
				return
			}
			logger.Info("created Kubernetes CronJob", "CronJob.Name", cronJob.Name, "CronJob.Namespace", cronJob.Namespace)
		}
	} else {
		cronJob.ResourceVersion = existingCronJob.ResourceVersion

		if err = k8sClient.Update(ctx, cronJob); err != nil {
			logger.Error(err, "failed to configure existing Kubernetes CronJob", "CronJob.Name", cronJob.Name, "CronJob.Namespace", cronJob.Namespace)
			return
		}
		logger.Info("configured Kubernetes CronJob", "CronJob.Name", cronJob.Name, "CronJob.Namespace", cronJob.Namespace)
	}

	if err = adapterutil.UpdateCwnpStatus(ctx, k8sClient, cronJob.Namespace+"/CronJob/"+cronJob.Name, cwnp.Name, false); err != nil {
		logger.Error(err, "failed to update ClusterNimbusPolicy status")
	}
}

func deleteCronJobs(ctx context.Context, logger logr.Logger, cwnpName string, cronJobsToDelete map[string]batchv1.CronJob) {
	for cronJobName := range cronJobsToDelete {
		cronJob := cronJobsToDelete[cronJobName]
		if err := k8sClient.Delete(ctx, &cronJob); err != nil {
			logger.Error(err, "failed to delete Kubernetes CronJob", "CronJob.Name", cronJobName, "CronJob.Namespace", cronJob.Namespace)
			continue
		}

		if err := adapterutil.UpdateCwnpStatus(ctx, k8sClient, cronJob.Namespace+"/CronJob/"+cronJob.Name, cwnpName, true); err != nil {
			logger.Error(err, "failed to update ClusterNimbusPolicy status")
		}
		logger.Info("Dangling Kubernetes CronJob deleted", "CronJobJob.Name", cronJob.Name, "CronJob.Namespace", cronJob.Namespace)
	}
}

func createCm(ctx context.Context, cwnp v1alpha1.ClusterNimbusPolicy, scheme *runtime.Scheme, k8sClient client.Client, configMap *corev1.ConfigMap) error {
	logger := log.FromContext(ctx)
	configMap.SetNamespace(K8tlsNamespace)
	if err := ctrl.SetControllerReference(&cwnp, configMap, scheme); err != nil {
		return err
	}

	var existingCm corev1.ConfigMap
	err := k8sClient.Get(ctx, client.ObjectKeyFromObject(configMap), &existingCm)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if err != nil {
		if errors.IsNotFound(err) {
			if err := k8sClient.Create(ctx, configMap); err != nil {
				return err
			}
			logger.Info(fmt.Sprintf("created configmap %s/%s", configMap.GetNamespace(), configMap.GetName()))
		}
	} else {
		configMap.SetResourceVersion(existingCm.GetResourceVersion())
		if err := k8sClient.Update(ctx, configMap); err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("configured configmap %s/%s", configMap.GetNamespace(), configMap.GetName()))
	}

	return nil
}
