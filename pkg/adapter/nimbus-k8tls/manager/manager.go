// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package manager

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/common"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-k8tls/builder"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-k8tls/watcher"
	adapterutil "github.com/5GSEC/nimbus/pkg/adapter/util"
	globalwatcher "github.com/5GSEC/nimbus/pkg/adapter/watcher"
)

var (
	scheme         = runtime.NewScheme()
	k8sClient      client.Client
	K8tlsNamespace = "nimbus-k8tls-env"
	k8tls          = "k8tls"
)

func init() {
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(batchv1.AddToScheme(scheme))
	utilruntime.Must(rbacv1.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	k8sClient = k8s.NewOrDie(scheme)
}

//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=clusternimbuspolicies,verbs=get;list;watch
//+kubebuilder:rbac:groups=intent.security.nimbus.com,resources=clusternimbuspolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;create;delete;list;watch;update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;create;delete;update
//+kubebuilder:rbac:groups="",resources=namespaces;serviceaccounts,verbs=get

func Run(ctx context.Context) {
	cwnpCh := make(chan string)
	deletedcwnpCh := make(chan *unstructured.Unstructured)
	go globalwatcher.WatchClusterNimbusPolicies(ctx, cwnpCh, deletedcwnpCh)

	updateCronJobCh := make(chan common.Request)
	deletedCronJobCh := make(chan common.Request)
	go watcher.WatchCronJobs(ctx, updateCronJobCh, deletedCronJobCh)

	// Get the namespace name within which the k8tls environment needs to be set
	for {
		select {
		case <-ctx.Done():
			close(cwnpCh)
			close(deletedcwnpCh)

			close(updateCronJobCh)
			close(deletedCronJobCh)
			return

		case createdCwnp := <-cwnpCh:
			createOrUpdateCronJob(ctx, createdCwnp)
		case deletedCwnp := <-deletedcwnpCh:
			logCronJobsToDelete(ctx, deletedCwnp)

		case updatedCronJob := <-updateCronJobCh:
			reconcileCronJob(ctx, updatedCronJob.Name)
		case deletedCronJob := <-deletedCronJobCh:
			reconcileCronJob(ctx, deletedCronJob.Name)
		}
	}
}

func reconcileCronJob(ctx context.Context, name string) {
	logger := log.FromContext(ctx)
	cwnpName := adapterutil.ExtractAnyNimbusPolicyName(name)
	var cwnp v1alpha1.ClusterNimbusPolicy
	err := k8sClient.Get(ctx, types.NamespacedName{Name: cwnpName}, &cwnp)
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "failed to get ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", cwnpName)
		}
		return
	}
	createOrUpdateCronJob(ctx, cwnpName)
}

func createOrUpdateCronJob(ctx context.Context, cwnpName string) {
	logger := log.FromContext(ctx)
	var cwnp v1alpha1.ClusterNimbusPolicy
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: cwnpName}, &cwnp); err != nil {
		logger.Error(err, "failed to get ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", cwnpName)
		return
	}

	if adapterutil.IsOrphan(cwnp.GetOwnerReferences(), "ClusterSecurityIntentBinding") {
		logger.V(4).Info("Ignoring orphan ClusterNimbusPolicy", "ClusterNimbusPolicy.Name", cwnpName)
		return
	}

	deleteDanglingCj(ctx, logger, cwnp)
	newCtx := context.WithValue(ctx, common.K8sClientKey, k8sClient)
	newCtx = context.WithValue(newCtx, common.NamespaceNameKey, K8tlsNamespace)
	cronJob, configMap := builder.BuildCronJob(newCtx, cwnp)

	if cronJob != nil {
		if !k8tlsEnvExist(ctx, k8sClient) {
			return
		}

		if configMap != nil {
			if err := createCm(ctx, cwnp, scheme, k8sClient, configMap); err != nil {
				logger.Error(err, "failed to create ConfigMap", "ConfigMap.Name", configMap.Name)
				return
			}
		}
		createOrUpdateCj(ctx, logger, cwnp, cronJob)
	}
}

func logCronJobsToDelete(ctx context.Context, deletedCwnp *unstructured.Unstructured) {
	logger := log.FromContext(ctx)

	var existingCronJobs batchv1.CronJobList
	if err := k8sClient.List(ctx, &existingCronJobs, &client.ListOptions{Namespace: K8tlsNamespace}); err != nil {
		logger.Error(err, "failed to list Kubernetes CronJob")
		return
	}

	for _, cronJob := range existingCronJobs.Items {
		for _, ownerRef := range cronJob.OwnerReferences {
			if ownerRef.Name == deletedCwnp.GetName() && ownerRef.UID == deletedCwnp.GetUID() {
				logger.Info(
					"Kubernetes CronJob deleted due to ClusterNimbusPolicy deletion",
					"CronJob.Name", cronJob.Name, "CronJob.Namespace", cronJob.Namespace,
					"ClusterNimbusPolicy.Name", deletedCwnp.GetName(),
				)
				break
			}
		}
	}
}

func deleteDanglingCj(ctx context.Context, logger logr.Logger, cwnp v1alpha1.ClusterNimbusPolicy) {
	var existingCronJobs batchv1.CronJobList
	if err := k8sClient.List(ctx, &existingCronJobs, &client.ListOptions{Namespace: K8tlsNamespace}); err != nil {
		logger.Error(err, "failed to list Kubernetes CronJob for cleanup")
		return
	}

	cronJobsToDelete := make(map[string]batchv1.CronJob)
	for _, cronJob := range existingCronJobs.Items {
		for _, ownerRef := range cronJob.OwnerReferences {
			if ownerRef.Name == cwnp.GetName() && ownerRef.UID == cwnp.GetUID() {
				cronJobsToDelete[cronJob.Name] = cronJob
				break
			}
		}
	}

	if len(cronJobsToDelete) == 0 {
		return
	}

	for _, nimbusRule := range cwnp.Spec.NimbusRules {
		cjName := cwnp.Name + "-" + strings.ToLower(nimbusRule.ID)
		delete(cronJobsToDelete, cjName)
	}

	deleteCronJobs(ctx, logger, cwnp.Name, cronJobsToDelete)
}
