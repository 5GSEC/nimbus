// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	v1alpha1 "github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-kyverno/utils"
	"github.com/go-logr/logr"
	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v2"
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
	background := true
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
				kp.Spec.Background = &background
				
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
		kps = append(kps, escapeToHost(np))
	case idpool.CocoWorkload:
		kpols, err := cocoRuntimeAddition(np)
		if err != nil {
			return kps, err
		}
		kps = append(kps, kpols...)
	case idpool.VirtualPatch:
		kpols, err := virtualPatch(np)
		if err != nil {
			return kps, err
		}
		kps = append(kps, kpols...)
	}
	return kps, nil
}

func escapeToHost(np *v1alpha1.NimbusPolicy) kyvernov1.Policy {
	rule := np.Spec.NimbusRules[0].Rule
	var psa_level api.Level = api.LevelBaseline
	var matchResourceFilters []kyvernov1.ResourceFilter

	if rule.Params["psa_level"] != nil {

		switch rule.Params["psa_level"][0] {
		case "restricted":
			psa_level = api.LevelRestricted

		default:
			psa_level = api.LevelBaseline
		}
	}

	labels := np.Spec.Selector.MatchLabels

	if len(labels) > 0 {
		for key, value := range labels {
			resourceFilter := kyvernov1.ResourceFilter {
				ResourceDescription: kyvernov1.ResourceDescription{
					Kinds: []string{
						"v1/Pod",
					},
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							key: value,
						},
					},
				},
			}
			matchResourceFilters = append(matchResourceFilters, resourceFilter)
		}
	} else {
		resourceFilter := kyvernov1.ResourceFilter{
			ResourceDescription: kyvernov1.ResourceDescription{
				Kinds: []string{
					"v1/Pod",
				},
			},
		}
		matchResourceFilters = append(matchResourceFilters, resourceFilter)
	}

	background := true
	kp := kyvernov1.Policy{
		Spec: kyvernov1.Spec{
			Background: &background,
			Rules: []kyvernov1.Rule{
				{
					Name: "pod-security-standard",
					MatchResources: kyvernov1.MatchResources{
						Any: matchResourceFilters,
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

	return kp
}

func cocoRuntimeAddition(np *v1alpha1.NimbusPolicy) ([]kyvernov1.Policy, error) {
	var kps []kyvernov1.Policy
	var errs []error
	var deployNames []string
	var mutateTargetResourceSpecs []kyvernov1.TargetResourceSpec
	var matchResourceFilters []kyvernov1.ResourceFilter
	labels := np.Spec.Selector.MatchLabels
	runtimeClass := "kata-clh"
	params := np.Spec.NimbusRules[0].Rule.Params["runtimeClass"]
	if params != nil {
		runtimeClass = params[0] 
	}
	patchStrategicMerge := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"runtimeClassName": runtimeClass,
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
	var markLabels = make(map[string][]string)
	for _, d := range deployments.Items {
		for k, v := range d.GetLabels() {
			key := k + ":" + v
			markLabels[key] = append(markLabels[key], d.GetName())
		}
	}
	for k, v := range labels {
		key := k + ":" + v
		if len(markLabels[key]) != 0 {
			deployNames = append(deployNames, markLabels[key]...)
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
	if len(labels) > 0 {
		for key, value := range labels {
			resourceFilter := kyvernov1.ResourceFilter{
				ResourceDescription: kyvernov1.ResourceDescription{
					Kinds: []string{
						"apps/v1/Deployment",
					},
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							key: value,
						},
					},
				},
			}
			matchResourceFilters = append(matchResourceFilters, resourceFilter)
		}
	} else if len(labels) == 0 {
		mutateResourceSpec := kyvernov1.TargetResourceSpec{
			ResourceSpec: kyvernov1.ResourceSpec{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
		}
		mutateTargetResourceSpecs = append(mutateTargetResourceSpecs, mutateResourceSpec)

		resourceFilter := kyvernov1.ResourceFilter{
			ResourceDescription: kyvernov1.ResourceDescription{
				Kinds: []string{
					"apps/v1/Deployment",
				},
			},
		}

		matchResourceFilters = append(matchResourceFilters, resourceFilter)
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
										"intent.security.nimbus.com/v1alpha1/NimbusPolicy",
									},
									Name: np.Name,
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

	mutateNewKp := kyvernov1.Policy{
		Spec: kyvernov1.Spec{

			Rules: []kyvernov1.Rule{
				{
					Name: "add runtime",
					MatchResources: kyvernov1.MatchResources{
						Any: matchResourceFilters,
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

	mutateNewKp.Name = np.Name + "-mutateoncreate"

	if (len(deployNames) > 0) || (len(labels) == 0 && len(deployments.Items) > 0) { // if labels are present but no deploy exists with matching label or labels are not present but deployments exists
		kps = append(kps, mutateExistingKp)
	}
	kps = append(kps, mutateNewKp)

	if len(errs) != 0 {
		return kps, nil
	}
	return kps, multierr.Combine(errs...)
}

func virtualPatch(np *v1alpha1.NimbusPolicy) ([]kyvernov1.Policy, error) {
	rule := np.Spec.NimbusRules[0].Rule
	requiredCVES := rule.Params["cve_list"]
	// cvePolicyMap := make(map[string]any)
	// schedule := rule.Params["schedule"][0]
	var kps []kyvernov1.Policy
	resp, err := utils.GetData[[]map[string]any]()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the response from knoxguard: %s", err.Error())
	}
	for _, currObj := range(resp) {
		image := currObj["image"].(string)
		fmt.Println(image)
		fmt.Println("------------------------------------------------------------------")
		cves := currObj["cves"].([]any)
		for _, obj := range(cves) {
			cveData := obj.(map[string]any)
			cve := cveData["cve"].(string)
			if utils.Contains(requiredCVES, cve) {
				fmt.Println(cveData["virtual_patch"])
				// create generate kyverno policies which will generate the native virtual patch policies based on the CVE's
				karmorPolCount := 1
				kyPolCount := 1
				netPolCount := 1
				virtual_patch := cveData["virtual_patch"].([]any)
				for _, policy := range(virtual_patch) {
					pol :=  policy.(map[string]any)
					policyData, ok := pol["karmor"].(map[string]any)
					if ok {
						kps = append(kps, generatePol("karmor", cve, image, np, policyData, karmorPolCount))
						karmorPolCount +=1
					}
				}
			}
		}
	}
	// Marshal the data into YAML
	// yamlData, err := yaml.Marshal(&resp)
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to parse the response to YAML: %s", err.Error())
	// }
	// Print the YAML data
	// fmt.Println(string(yamlData))
	return kps, nil
}

func addManagedByAnnotation(kp *kyvernov1.Policy) {
	kp.Annotations["app.kubernetes.io/managed-by"] = "nimbus-kyverno"
}

func generatePol(polengine string, cve string, image string, np *v1alpha1.NimbusPolicy, policyData map[string]any, count int) (kyvernov1.Policy) {
	var pol kyvernov1.Policy
	labels := np.Spec.Selector.MatchLabels
	cve  = strings.ToLower(cve)
	// Marshal the data into YAML
	yamlData, err := yaml.Marshal(&policyData)
	if err != nil {
		fmt.Println("unable to parse the response to YAML: ", err.Error()) 
		return pol
	}

	preconditionJSON := `{
        "all": [
            {
                "key": "${ image }",
                "operator": "AnyIn",
                "value": "{{ images.containers.*.name }}"
            }
        ]
    }`

	// Convert JSON string to map
    var preConditionMap map[string]interface{}
    if err := json.Unmarshal([]byte(preconditionJSON), &preConditionMap); err != nil {
        panic(err)
    }

	// Convert map to apiextensions.JSON
    apiextensionsJSON := v1.JSON{}
    apiextensionsJSON.Raw, _ = json.MarshalIndent(preConditionMap, "", " ")

	if polengine == "karmor" {
		pol = kyvernov1.Policy {
			ObjectMeta: metav1.ObjectMeta{
				Name: np.Name + cve + string(count),
			},
			Spec: kyvernov1.Spec{
				GenerateExisting: true,
				Rules: []kyvernov1.Rule{
					{
						Name: cve+"virtual-patch-karmor",
						MatchResources: kyvernov1.MatchResources{
							Any: kyvernov1.ResourceFilters{
								{
									ResourceDescription: kyvernov1.ResourceDescription{
										Kinds: []string{
											"v1/Pod",
										},
										Selector: &metav1.LabelSelector{
											MatchLabels: labels,
										},
									},
								},
							},
						},
						RawAnyAllConditions: kyvernov1.ToJSON(apiextensionsJSON.Raw),
						
					},
				},
			},
		}
	}
	// Print the YAML data
	fmt.Println(string(yamlData))
	return pol
}