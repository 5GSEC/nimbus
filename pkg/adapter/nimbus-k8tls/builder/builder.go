// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package builder

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/common"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
)

var (
	DefaultSchedule = "@weekly"
	backOffLimit    = int32(5)
)

func BuildCronJob(ctx context.Context, cwnp v1alpha1.ClusterNimbusPolicy) (*batchv1.CronJob, *corev1.ConfigMap) {
	logger := log.FromContext(ctx)
	for _, nimbusRule := range cwnp.Spec.NimbusRules {
		id := nimbusRule.ID
		if idpool.IsIdSupportedBy(id, "k8tls") {
			cronJob, configMap := cronJobFor(ctx, id, nimbusRule)
			cronJob.SetName(cwnp.Name + "-" + strings.ToLower(id))
			cronJob.SetAnnotations(map[string]string{
				"app.kubernetes.io/managed-by": "nimbus-k8tls",
			})
			cronJob.SetLabels(cwnp.Labels)
			return cronJob, configMap
		}
		logger.Info("K8TLS adapter doesn't support this ID", "ID", id)
	}
	return nil, nil
}

func cronJobFor(ctx context.Context, id string, rule v1alpha1.NimbusRules) (*batchv1.CronJob, *corev1.ConfigMap) {
	switch id {
	case idpool.AssessTLS:
		return assessTlsCronJob(ctx, rule)
	default:
		return nil, nil
	}
}

func assessTlsCronJob(ctx context.Context, rule v1alpha1.NimbusRules) (*batchv1.CronJob, *corev1.ConfigMap) {
	schedule, scheduleKeyExists := rule.Rule.Params["schedule"]
	externalAddresses, addrKeyExists := rule.Rule.Params["external_addresses"]
	if scheduleKeyExists && addrKeyExists {
		return cronJobForAssessTls(ctx, schedule[0], externalAddresses...)
	}
	if scheduleKeyExists {
		return cronJobForAssessTls(ctx, schedule[0])
	}
	if addrKeyExists {
		return cronJobForAssessTls(ctx, DefaultSchedule, externalAddresses...)
	}
	return cronJobForAssessTls(ctx, DefaultSchedule)
}

func cronJobForAssessTls(ctx context.Context, schedule string, externalAddresses ...string) (*batchv1.CronJob, *corev1.ConfigMap) {
	logger := log.FromContext(ctx)
	cj := &batchv1.CronJob{
		Spec: batchv1.CronJobSpec{
			Schedule: schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					BackoffLimit: &backOffLimit,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							InitContainers: []corev1.Container{
								{
									Name:            "k8tls",
									Image:           "kubearmor/k8tls:latest",
									Command:         []string{"./k8s_tlsscan"},
									ImagePullPolicy: corev1.PullAlways,
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "fips-config",
											MountPath: "/home/k8tls/config/",
											ReadOnly:  true,
										},
										{
											Name:      "k8tls-report",
											MountPath: "/tmp/",
										},
									},
								},
							},
							Containers: []corev1.Container{
								{
									Name:            "fluent-bit",
									Image:           "fluent/fluent-bit:latest",
									ImagePullPolicy: corev1.PullAlways,
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "fluent-bit-config",
											MountPath: "/fluent-bit/etc/fluent-bit.conf",
											SubPath:   "fluent-bit.conf",
											ReadOnly:  true,
										},
										{
											Name:      "k8tls-report",
											MountPath: "/tmp/",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "fips-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "fips-config",
											},
										},
									},
								},
								{
									Name: "fluent-bit-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "fluent-bit-config",
											},
										},
									},
								},
								{
									Name: "k8tls-report",
									VolumeSource: corev1.VolumeSource{
										EmptyDir: &corev1.EmptyDirVolumeSource{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Fetch the elasticsearch password secret. If the secret is present, set TTLSecondsAfterFinished and reference the secret in the cronjob templateZ
	var elasticsearchPasswordSecret corev1.Secret
	err := ctx.Value(common.K8sClientKey).(client.Client).Get(ctx, client.ObjectKey{Namespace: ctx.Value(common.NamespaceNameKey).(string), Name: "elasticsearch-password"}, &elasticsearchPasswordSecret)
	if err == nil {
		// Convert string to int
		i, err := strconv.ParseInt(os.Getenv("TTLSECONDSAFTERFINISHED"), 10, 32)
		if err != nil {
			logger.Error(err, "Error converting string to int", "TTLSECONDSAFTERFINISHED: ", os.Getenv("TTLSECONDSAFTERFINISHED"))
			return nil, nil
		}
		// Convert int to int32
		ttlSecondsAfterFinished := int32(i)
		// If we are sending the report to elasticsearch, then we delete the pod spawned by job after 1 hour. Else we keep the pod
		cj.Spec.JobTemplate.Spec.TTLSecondsAfterFinished = &ttlSecondsAfterFinished
		cj.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{
			{
				Name: "ES_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "elasticsearch-password",
						},
						Key: "es_password",
					},
				},
			},
		}
	}

	if len(externalAddresses) > 0 {
		cm := buildConfigMap(externalAddresses)

		cj.Spec.JobTemplate.Spec.Template.Spec.InitContainers[0].VolumeMounts = append(cj.Spec.JobTemplate.Spec.Template.Spec.InitContainers[0].VolumeMounts, corev1.VolumeMount{
			Name:      cm.Name,
			ReadOnly:  true,
			MountPath: "/var/k8tls/",
		})
		cj.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(cj.Spec.JobTemplate.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: cm.Name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cm.Name,
					},
				},
			},
		})

		cj.Spec.JobTemplate.Spec.Template.Spec.InitContainers[0].Command[0] = "./tlsscan"
		cj.Spec.JobTemplate.Spec.Template.Spec.InitContainers[0].Command = append(cj.Spec.JobTemplate.Spec.Template.Spec.InitContainers[0].Command,
			"--infile",
			cj.Spec.JobTemplate.Spec.Template.Spec.InitContainers[0].VolumeMounts[2].MountPath+"addresses",
			"--compact-json",
		)
		return cj, cm
	}

	return cj, nil
}

func buildConfigMap(externalAddresses []string) *corev1.ConfigMap {
	addresses := formatAddresses(externalAddresses)
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "external-addresses",
		},
		Data: map[string]string{
			"addresses": strings.Join(addresses, "\n"),
		},
	}
}

func formatAddresses(externalAddresses []string) []string {
	addresses := make([]string, 0, len(externalAddresses))
	for _, externalAddress := range externalAddresses {
		domain := strings.Split(strings.Split(externalAddress, ":")[0], ".")[0]
		domainTitle := strings.ToUpper(string(domain[0])) + domain[1:]
		addresses = append(addresses, fmt.Sprintf("%s %s", externalAddress, domainTitle))
	}
	return addresses
}
