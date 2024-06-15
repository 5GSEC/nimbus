// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"

	v1 "github.com/5GSEC/nimbus/api/v1"
	common "github.com/5GSEC/nimbus/pkg/adapter/common"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
)

func BuildpodsFromCoco(logger logr.Logger, np *v1.NimbusPolicy, oldPod *corev1.Pod) []corev1.Pod {
	// Build pods based on given IDs
	var pods []corev1.Pod
	for _, nimbusRule := range np.Spec.NimbusRules {
		id := nimbusRule.ID
		if idpool.IsIdSupportedBy(id, "coco") {
			pod := buildPodFor(id, oldPod)
			pod.Name = fmt.Sprintf("%s-cvm", oldPod.Name)
			pod.Namespace = np.Namespace
			pod.ObjectMeta.Labels = np.Spec.Selector.MatchLabels
			AddManagedByAnnotationPod(&pod)
			pods = append(pods, pod)
		} else {
			logger.Info("Coco adapter does not support this ID", "ID", id,
				"NimbusPolicy.Name", np.Name, "NimbusPolicy.Namespace", np.Namespace)
		}
	}
	return pods
}

func buildPodFor(id string, oldPod *corev1.Pod) corev1.Pod {
	switch id {
	case idpool.CocoWorkload:
		return cocoWorkloadPod(oldPod)
	default:
		return corev1.Pod{}
	}
}

func cocoWorkloadPod(oldPod *corev1.Pod) corev1.Pod {
	runtimeClassName := "kata-qemu-snp"

	return corev1.Pod{
		Spec: corev1.PodSpec{
			RuntimeClassName: &runtimeClassName,
			Containers:       oldPod.Spec.Containers,
			ImagePullSecrets: oldPod.Spec.ImagePullSecrets,
			Volumes:          oldPod.Spec.Volumes,
		},
	}
}

func BuildpodsFromK8s(logger logr.Logger, podData common.PodData) corev1.Pod {
	pod := normalPod(podData)
	pod.Name = removeIDPrefix(podData.Name)
	pod.Namespace = podData.Namespace
	return pod
}

func normalPod(podData common.PodData) corev1.Pod {
	return corev1.Pod{
		Spec: corev1.PodSpec{
			Containers:       podData.Spec.Containers,
			ImagePullSecrets: podData.Spec.ImagePullSecrets,
			Volumes:          podData.Spec.Volumes,
		},
	}
}

func removeIDPrefix(podName string) string {
	suffix := "-cvm"
	if strings.HasSuffix(podName, suffix) {
		return podName[:len(podName)-len(suffix)]
	}
	return podName
}

func AddManagedByAnnotationPod(pod *corev1.Pod) {
	pod.Annotations = make(map[string]string)
	pod.Annotations["app.kubernetes.io/managed-by"] = "nimbus-coco"
}
