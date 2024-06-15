// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package manager

import (
	"context"

	intentv1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func isNonCVMDeploy(deployment *appsv1.Deployment) bool {
	return deployment.Spec.Template.Spec.RuntimeClassName == nil || *deployment.Spec.Template.Spec.RuntimeClassName != "kata-qemu-snp"
}

func isRunningOnCVMDeploy(deployment *appsv1.Deployment) bool {
	return deployment.Spec.Template.Spec.RuntimeClassName != nil && *deployment.Spec.Template.Spec.RuntimeClassName == "kata-qemu-snp"
}

func isNonCVMPod(pod *corev1.Pod) bool {
	return pod.Spec.RuntimeClassName == nil || *pod.Spec.RuntimeClassName != "kata-qemu-snp"
}

func isRunningOnCVMPod(pod *corev1.Pod) bool {
	return pod.Spec.RuntimeClassName != nil && *pod.Spec.RuntimeClassName == "kata-qemu-snp"
}

func checkLabelMatch(Labels, policyLabels map[string]string) bool {
	for k, v := range policyLabels {
		if Labels[k] != v {
			return false
		}
	}
	return true
}

func listDeploy(ctx context.Context, selector map[string]string) ([]appsv1.Deployment, error) {
	var deploymentList appsv1.DeploymentList
	listOpts := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(selector),
	}
	if err := k8sClient.List(ctx, &deploymentList, listOpts); err != nil {
		return nil, err
	}
	return deploymentList.Items, nil
}

func listPodsBySelector(ctx context.Context, selector map[string]string) ([]corev1.Pod, error) {
	var podList corev1.PodList
	listOpts := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(selector),
	}
	if err := k8sClient.List(ctx, &podList, listOpts); err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func getNP(ctx context.Context, logger logr.Logger, npName, namespace string) (*intentv1.NimbusPolicy, error) {
	var np intentv1.NimbusPolicy

	err := k8sClient.Get(ctx, types.NamespacedName{Name: npName, Namespace: namespace}, &np)
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "failed to get NimbusPolicy", "NimbusPolicy.Name", npName, "NimbusPolicy.Namespace", namespace)
		}
		return nil, err
	}

	return &np, nil
}

func getPod(ctx context.Context, podName, namespace string) (*corev1.Pod, error) {
	var pod corev1.Pod
	err := k8sClient.Get(ctx, types.NamespacedName{Name: podName, Namespace: namespace}, &pod)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}

func getDeployFromPod(ctx context.Context, pod *corev1.Pod) (*appsv1.Deployment, error) {
	var deploymentList appsv1.DeploymentList
	listOpts := &client.ListOptions{
		Namespace: pod.Namespace,
	}
	if err := k8sClient.List(ctx, &deploymentList, listOpts); err != nil {
		return nil, err
	}

	for _, deployment := range deploymentList.Items {
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == "ReplicaSet" {
				rsName := owner.Name
				rsNamespace := pod.Namespace

				var replicaSetList appsv1.ReplicaSetList
				rsListOpts := &client.ListOptions{
					Namespace: rsNamespace,
				}
				if err := k8sClient.List(ctx, &replicaSetList, rsListOpts); err != nil {
					return nil, err
				}

				for _, rs := range replicaSetList.Items {
					if rs.Name == rsName && rs.OwnerReferences[0].Kind == "Deployment" && rs.OwnerReferences[0].Name == deployment.Name {
						return &deployment, nil
					}
				}
			}
		}
	}

	return nil, errors.NewNotFound(appsv1.Resource("deployments"), pod.Name)
}
