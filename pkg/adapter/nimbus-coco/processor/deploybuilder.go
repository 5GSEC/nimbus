// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/5GSEC/nimbus/api/v1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
)

func BuildDeployFromCVM(logger logr.Logger, np *v1.NimbusPolicy, oldDeployment *appsv1.Deployment) []appsv1.Deployment {
	// Build deployments based on given IDs
	var deployments []appsv1.Deployment
	for _, nimbusRule := range np.Spec.NimbusRules {
		id := nimbusRule.ID
		if idpool.IsIdSupportedBy(id, "coco") {
			deployment := buildDeployFor(id, oldDeployment, np)
			deployment.Name = oldDeployment.Name
			deployment.Namespace = np.Namespace
			deployment.Spec.Template.ObjectMeta.Labels = np.Spec.Selector.MatchLabels
			AddManagedByAnnotation(&deployment)
			deployments = append(deployments, deployment)
		} else {
			logger.Info("Coco adapter does not support this ID", "ID", id,
				"NimbusPolicy.Name", np.Name, "NimbusPolicy.Namespace", np.Namespace)
		}
	}
	return deployments
}

func buildDeployFor(id string, oldDeployment *appsv1.Deployment, np *v1.NimbusPolicy) appsv1.Deployment {
	switch id {
	case idpool.CocoWorkload:
		return cocoWorkloadDeploy(oldDeployment, np)
	default:
		return appsv1.Deployment{}
	}
}

func cocoWorkloadDeploy(oldDeployment *appsv1.Deployment, np *v1.NimbusPolicy) appsv1.Deployment {
	runtimeClassName := "kata-qemu-snp"
	replicas := int32(1) // Adjust as needed

	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      oldDeployment.Name,
			Namespace: np.Namespace,
			Labels:    np.Spec.Selector.MatchLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: np.Spec.Selector.MatchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: np.Spec.Selector.MatchLabels,
				},
				Spec: corev1.PodSpec{
					RuntimeClassName: &runtimeClassName,
					Containers:       oldDeployment.Spec.Template.Spec.Containers,
					ImagePullSecrets: oldDeployment.Spec.Template.Spec.ImagePullSecrets,
					Volumes:          oldDeployment.Spec.Template.Spec.Volumes,
				},
			},
		},
	}
}

func BuildDeployFromPod(pod *corev1.Pod, np *v1.NimbusPolicy) appsv1.Deployment {
	replicas := int32(1)
	runtimeClassName := "kata-qemu-snp"

	newDeployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Labels:    pod.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: pod.Labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: pod.Labels,
				},
				Spec: corev1.PodSpec{
					RuntimeClassName: &runtimeClassName,
					Containers:       pod.Spec.Containers,
					ImagePullSecrets: pod.Spec.ImagePullSecrets,
					Volumes:          pod.Spec.Volumes,
				},
			},
		},
	}

	AddManagedByAnnotation(&newDeployment)
	return newDeployment
}

func BuildDeployFromK8s(logger logr.Logger, oldDeployment appsv1.Deployment) appsv1.Deployment {
	deployment := normalPodDeploy(oldDeployment)
	deployment.Name = oldDeployment.Name
	deployment.Namespace = oldDeployment.Namespace
	deployment.Spec.Template.ObjectMeta.Labels = oldDeployment.Spec.Template.Labels
	AddManagedByAnnotation(&deployment)
	return deployment
}

func normalPodDeploy(deployment appsv1.Deployment) appsv1.Deployment {
	replicas := int32(1) // Adjust as needed
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: deployment.Spec.Selector.MatchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: deployment.Spec.Template.Labels,
				},
				Spec: corev1.PodSpec{
					Containers:       deployment.Spec.Template.Spec.Containers,
					ImagePullSecrets: deployment.Spec.Template.Spec.ImagePullSecrets,
					Volumes:          deployment.Spec.Template.Spec.Volumes,
				},
			},
		},
	}
}

func AddManagedByAnnotation(deployment *appsv1.Deployment) {
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}
	deployment.Annotations["app.kubernetes.io/managed-by"] = "nimbus-coco"
}
