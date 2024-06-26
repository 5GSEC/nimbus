// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"context"
	"encoding/json"
	"strings"

	v1alpha1 "github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	"github.com/go-logr/logr"
	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	"go.uber.org/multierr"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/pod-security-admission/api"
)

var (
	client dynamic.Interface
)

func init() {
	client = k8s.NewDynamicClient()
}

func BuildKpsFrom(logger logr.Logger, np *v1alpha1.NimbusPolicy) []kyvernov1.Policy {
	// Build KPs based on given IDs
	var allkps []kyvernov1.Policy
	for _, nimbusRule := range np.Spec.NimbusRules {
		id := nimbusRule.ID
		if idpool.IsIdSupportedBy(id, "kyverno") {
			kps, err := buildKpFor(id, np)
			if err != nil {
				logger.Error(err, "error while building kyverno policies")
			}
			for _, kp := range kps {
				if id != "cocoWorkload" {
					kp.Name = np.Name + "-" + strings.ToLower(id)
				}
				kp.Namespace = np.Namespace
				kp.Annotations = make(map[string]string)
				kp.Annotations["policies.kyverno.io/description"] = nimbusRule.Description
				if nimbusRule.Rule.RuleAction == "Block" {
					kp.Spec.ValidationFailureAction = kyvernov1.ValidationFailureAction("Enforce")
				} else {
					kp.Spec.ValidationFailureAction = kyvernov1.ValidationFailureAction("Audit")
				}
				addManagedByAnnotation(&kp)
				allkps = append(allkps, kp)
			}
		} else {
			logger.Info("Kyverno does not support this ID", "ID", id,
				"NimbusPolicy", np.Name, "NimbusPolicy.Namespace", np.Namespace)
		}
	}
	return allkps
}

// buildKpFor builds a KyvernoPolicy based on intent ID supported by Kyverno Policy Engine.
func buildKpFor(id string, np *v1alpha1.NimbusPolicy) ([]kyvernov1.Policy, error) {
	var kps []kyvernov1.Policy
	switch id {
	case idpool.EscapeToHost:
		kps = append(kps, escapeToHost(np, np.Spec.NimbusRules[0].Rule))
	case idpool.CocoWorkload:
		kpols, err := cocoRuntimeAddition(np, np.Spec.NimbusRules[0].Rule)
		if err != nil {
			return kps, err
		}
		kps = append(kps, kpols...)
	}
	return kps, nil
}

func cocoRuntimeAddition(np *v1alpha1.NimbusPolicy, rule v1alpha1.Rule) ([]kyvernov1.Policy, error) {
	var kps []kyvernov1.Policy
	var errs []error
	var deployNames []string
	var mutateTargetResourceSpecs []kyvernov1.TargetResourceSpec
	labels := np.Spec.Selector.MatchLabels
	patchStrategicMerge := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"runtimeClassName": "kata-clh",
				},
			},
		},
	}
	patchBytes, err := json.Marshal(patchStrategicMerge)
	if err != nil {
		errs = append(errs, err)
	}
	if err != nil {
		errs = append(errs, err)
	}

	deploymentsGVR := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	deployments, err := client.Resource(deploymentsGVR).Namespace(np.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		errs = append(errs, err)
	}

	for _, d := range deployments.Items {
		for key1, value1 := range d.GetLabels() {
			for key2, value2 := range np.Spec.Selector.MatchLabels {
				if key1 == key2 && value1 == value2 {
					deployNames = append(deployNames, d.GetName())
				}
			}
		}
	}

	for _, deployName := range deployNames {
		mutateResourceSpec := kyvernov1.TargetResourceSpec{
			ResourceSpec: kyvernov1.ResourceSpec{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       deployName,
			},
		}
		mutateTargetResourceSpecs = append(mutateTargetResourceSpecs, mutateResourceSpec)
	}

	if len(labels) == 0 {
		mutateResourceSpec := kyvernov1.TargetResourceSpec{
			ResourceSpec: kyvernov1.ResourceSpec{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
		}
		mutateTargetResourceSpecs = append(mutateTargetResourceSpecs, mutateResourceSpec)
	}

	mutateExistingKp := kyvernov1.Policy{
		Spec: kyvernov1.Spec{
			MutateExistingOnPolicyUpdate: true,
			Rules: []kyvernov1.Rule{
				{
					Name: "add runtime",
					MatchResources: kyvernov1.MatchResources{
						Any: kyvernov1.ResourceFilters{
							kyvernov1.ResourceFilter{
								ResourceDescription: kyvernov1.ResourceDescription{
									Kinds: []string{
										"v1/ConfigMap",
									},
									Name: np.Name+"-mutateexisting-trigger-configmap",
								},
							},
						},
					},
					Mutation: kyvernov1.Mutation{
						Targets: mutateTargetResourceSpecs,
						RawPatchStrategicMerge: &v1.JSON{
							Raw: patchBytes,
						},
					},
				},
			},
		},
	}
	mutateExistingKp.Name = np.Name + "-mutateexisting"

	mutateNonExistingKp := kyvernov1.Policy{
		Spec: kyvernov1.Spec{

			Rules: []kyvernov1.Rule{
				{
					Name: "add runtime",
					MatchResources: kyvernov1.MatchResources{
						Any: kyvernov1.ResourceFilters{
							kyvernov1.ResourceFilter{
								ResourceDescription: kyvernov1.ResourceDescription{
									Kinds: []string{
										"apps/v1/Deployment",
									},
								},
							},
						},
					},
					Mutation: kyvernov1.Mutation{
						RawPatchStrategicMerge: &v1.JSON{
							Raw: patchBytes,
						},
					},
				},
			},
		},
	}

	if len(labels) > 0 {
		mutateNonExistingKp.Spec.Rules[0].MatchResources.Any[0].ResourceDescription.Selector = &metav1.LabelSelector{MatchLabels: labels}
	}
	mutateNonExistingKp.Name = np.Name + "-mutateoncreate"

	if len(deployNames) > 0 {  // if labels are present but no deploy exists with matching label
		kps = append(kps, mutateExistingKp)
	}
	kps = append(kps, mutateNonExistingKp)

	if len(errs) != 0 {
		return kps, nil
	}
	return kps, multierr.Combine(errs...)
}

func escapeToHost(np *v1alpha1.NimbusPolicy, rule v1alpha1.Rule) kyvernov1.Policy {

	var psa_level api.Level = api.LevelBaseline

	if rule.Params["psa_level"] != nil {

		switch rule.Params["psa_level"][0] {
		case "restricted":
			psa_level = api.LevelRestricted

		default:
			psa_level = api.LevelBaseline
		}
	}

	labels := np.Spec.Selector.MatchLabels

	background := true
	kp := kyvernov1.Policy{
		Spec: kyvernov1.Spec{
			Background: &background,
			Rules: []kyvernov1.Rule{
				{
					Name: "pod-security-standard",
					MatchResources: kyvernov1.MatchResources{
						Any: kyvernov1.ResourceFilters{
							kyvernov1.ResourceFilter{
								ResourceDescription: kyvernov1.ResourceDescription{
									Kinds: []string{
										"v1/Pod",
									},
								},
							},
						},
					},
					Validation: kyvernov1.Validation{
						PodSecurity: &kyvernov1.PodSecurity{
							Level:   psa_level,
							Version: "latest",
						},
					},
				},
			},
		},
	}

	if len(labels) > 0 {
		kp.Spec.Rules[0].MatchResources.Any[0].ResourceDescription.Selector = &metav1.LabelSelector{MatchLabels: labels}
	}

	return kp
}

func addManagedByAnnotation(kp *kyvernov1.Policy) {
	kp.Annotations["app.kubernetes.io/managed-by"] = "nimbus-kyverno"
}
