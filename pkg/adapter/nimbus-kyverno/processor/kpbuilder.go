// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	v1alpha1 "github.com/5GSEC/nimbus/api/v1alpha1"
	"github.com/5GSEC/nimbus/pkg/adapter/idpool"
	"github.com/5GSEC/nimbus/pkg/adapter/k8s"
	"github.com/5GSEC/nimbus/pkg/adapter/nimbus-kyverno/utils"
	"github.com/go-logr/logr"
	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	"github.com/robfig/cron/v3"
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
	background := true
	for _, nimbusRule := range np.Spec.NimbusRules {
		id := nimbusRule.ID
		if idpool.IsIdSupportedBy(id, "kyverno") {
			kps, err := buildKpFor(id, np, logger)
			if err != nil {
				logger.Error(err, "error while building kyverno policies")
			}
			for _, kp := range kps {
				if id != "cocoWorkload" && id != "virtualPatch" {
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
func buildKpFor(id string, np *v1alpha1.NimbusPolicy, logger logr.Logger) ([]kyvernov1.Policy, error) {
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
		kpols, err := virtualPatch(np, logger)
		if err != nil {
			return kps, err
		}
		kps = append(kps, kpols...)
		watchCVES(np, logger)
	}
	return kps, nil
}

func watchCVES(np *v1alpha1.NimbusPolicy, logger logr.Logger) {
	rule := np.Spec.NimbusRules[0].Rule
	schedule := "0 0 * * *"
	if rule.Params["schedule"] != nil {
		schedule = rule.Params["schedule"][0]
	}
    // Schedule the deletion of the Nimbus policy
    c := cron.New()
    _, err := c.AddFunc(schedule, func() {
        logger.Info("Checking for CVE updates and updation of policies")
        err := deleteNimbusPolicy(np, logger)
        if err != nil {
            logger.Error(err, "error while updating policies")
        }
    })
    if err != nil {
        logger.Error(err, "error while adding the schedule to update policies")
    }
    c.Start()

}



func deleteNimbusPolicy(np *v1alpha1.NimbusPolicy, logger logr.Logger) error {
    nimbusPolicyGVR := schema.GroupVersionResource{Group: "intent.security.nimbus.com", Version: "v1alpha1", Resource: "nimbuspolicies"}
	err := client.Resource(nimbusPolicyGVR).Namespace(np.Namespace).Delete(context.TODO(), np.Name,metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete Nimbus Policy: %s", err.Error())
	}
	logger.Info("Nimbus policy deleted successfully")
    return nil
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
			resourceFilter := kyvernov1.ResourceFilter{
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

func virtualPatch(np *v1alpha1.NimbusPolicy, logger logr.Logger) ([]kyvernov1.Policy, error) {
	rule := np.Spec.NimbusRules[0].Rule
	requiredCVES := rule.Params["cve_list"]
	var kps []kyvernov1.Policy
	resp, err := utils.FetchVirtualPatchData[[]map[string]any]()
	if err != nil {
		return kps, err
	}
	for _, currObj := range resp {
		image := currObj["image"].(string)
		cves := currObj["cves"].([]any)
		for _, obj := range cves {
			cveData := obj.(map[string]any)
			cve := cveData["cve"].(string)
			if utils.Contains(requiredCVES, cve) {
				// create generate kyverno policies which will generate the native virtual patch policies based on the CVE's
				karmorPolCount := 1
				kyvPolCount := 1
				netPolCount := 1
				virtualPatch := cveData["virtual_patch"].([]any)
				for _, policy := range virtualPatch {
					pol := policy.(map[string]any)
					policyData, ok := pol["karmor"].(map[string]any)
					if ok {
						karmorPol, err := generatePol("karmor", cve, image, np, policyData, karmorPolCount, logger)
						if err != nil {
							logger.V(2).Error(err, "Error while  generating karmor policy")
						}
						kps = append(kps, karmorPol)
						karmorPolCount += 1
					}
					policyData, ok = pol["kyverno"].(map[string]any)
					if ok {
						kyvernoPol, err := generatePol("kyverno", cve, image, np, policyData, kyvPolCount, logger)
						if err != nil {
							logger.V(2).Error(err, "Error while  generating kyverno policy")
						}
						kps = append(kps, kyvernoPol)
						kyvPolCount += 1
					}
					
					policyData, ok = pol["netpol"].(map[string]any)
					if ok {
						netPol, err := generatePol("netpol", cve, image, np, policyData, netPolCount, logger)
						if err != nil {
							logger.V(2).Error(err, "Error while  generating network policy")
						}
						kps = append(kps, netPol)
						netPolCount += 1
					}
				}
			}
		}
	}
	return kps, nil
}

func addManagedByAnnotation(kp *kyvernov1.Policy) {
	kp.Annotations["app.kubernetes.io/managed-by"] = "nimbus-kyverno"
}

func generatePol(polengine string, cve string, image string, np *v1alpha1.NimbusPolicy, policyData map[string]any, count int, logger logr.Logger) (kyvernov1.Policy, error) {
	var pol kyvernov1.Policy
	labels := np.Spec.Selector.MatchLabels
	cve = strings.ToLower(cve)
	uid := np.ObjectMeta.GetUID()
	ownerShipList := []any{
		map[string]any{
			"apiVersion":         "intent.security.nimbus.com/v1alpha1",
			"blockOwnerDeletion": true,
			"controller":         true,
			"kind":               "NimbusPolicy",
			"name":               np.GetName(),
			"uid":                uid,
		},
	}

	preConditionMap := map[string]any{
		"all": []any{
			map[string]any{
				"key":  image,
				"operator": "AnyIn",
				"value": "{{ request.object.spec.containers[].image }}",
			},
		},
	}
	preconditionBytes, _ := json.Marshal(preConditionMap)
	

	getPodName := kyvernov1.ContextEntry{
		Name: "podName",
		Variable: &kyvernov1.Variable{
			JMESPath: "request.object.metadata.name",
		},
	}

	metadataMap := policyData["metadata"].(map[string]any)

	// set OwnerShipRef for the generatedPol

	metadataMap["ownerReferences"] = ownerShipList

	specMap := policyData["spec"].(map[string]any)

	jmesPathContainerNameQuery := "request.object.spec.containers[?(@.image=='" + image + "')].name | [0]"


	delete(policyData, "apiVersion")
	delete(policyData, "kind")

	generatorPolicyName := np.Name + "-" + cve + "-"+ polengine + "-" + strconv.Itoa(count)


	// kubearmor policy generation

	if polengine == "karmor" {
		generatedPolicyName := metadataMap["name"].(string) + "-{{ podName }}"
		selector := specMap["selector"].(map[string]any)
		delete(selector, "matchLabels")
		selectorLabels := make(map[string]any)
		for key, value := range labels {
			selectorLabels[key] = value
		}
		selectorLabels["kubearmor.io/container.name"] = "{{ containerName }}"
		selector["matchLabels"] = selectorLabels

		policyBytes, err := json.Marshal(policyData)
		if err != nil {
			return pol, err
		}
		pol = kyvernov1.Policy{
			ObjectMeta: metav1.ObjectMeta{
				Name: generatorPolicyName,
			},
			Spec: kyvernov1.Spec{
				GenerateExisting: true,
				Rules: []kyvernov1.Rule{
					{
						Name: cve + "virtual-patch-karmor",
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
						RawAnyAllConditions: &v1.JSON{Raw: preconditionBytes},
						Context: []kyvernov1.ContextEntry{
							{
								Name: "containerName",
								Variable: &kyvernov1.Variable{
									JMESPath: jmesPathContainerNameQuery,
								},
							},
							getPodName,
						},
						Generation: kyvernov1.Generation{
							ResourceSpec: kyvernov1.ResourceSpec{
								APIVersion: "security.kubearmor.com/v1",
								Kind:       "KubeArmorPolicy",
								Name:       generatedPolicyName,
								Namespace:  np.GetNamespace(),
							},
							RawData: &v1.JSON{Raw: policyBytes},
						},
					},
				},
			},
		}
	}

	// kyverno policy generation

	if polengine == "kyverno" {

		generatedPolicyName := metadataMap["name"].(string)
		selectorMap := map[string]any{
			"matchLabels": labels,
		}

		kindMap := map[string]any{
			"kinds": []any{
				"Pod",
			},
			"selector": selectorMap,
		}

		newMatchMap := map[string]any{
			"any": []any{
				map[string]any{
					"resources": kindMap,
				},
			},
		}
		rulesMap := specMap["rules"].([]any)
		rule := rulesMap[0].(map[string]any)

		// adding resources as Pod and ommitting all the incoming resource types
		delete(rule, "match")
		rule["match"] = newMatchMap

		// appending the image matching precondition to the existing preconditions
		preCndMap := rule["preconditions"].(map[string]any)
		conditionsList, ok := preCndMap["any"].([]any)
		if ok {
			preConditionMap["all"] = append(preConditionMap["all"].([]any), conditionsList...)
		}

		delete(rule, "preconditions")

		rule["preconditions"] = preConditionMap

		policyBytes, err := json.Marshal(policyData)
		if err != nil {
			return pol, err
		}

		pol = kyvernov1.Policy{
			ObjectMeta: metav1.ObjectMeta{
				Name: generatorPolicyName,
			},
			Spec: kyvernov1.Spec{
				GenerateExisting: true,
				Rules: []kyvernov1.Rule{
					{
						Name: cve + "-virtual-patch-kyverno",
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
						Generation: kyvernov1.Generation{
							ResourceSpec: kyvernov1.ResourceSpec{
								APIVersion: "kyverno.io/v1",
								Kind:       "Policy",
								Name:       generatedPolicyName,
								Namespace:  np.GetNamespace(),
							},
							RawData: &v1.JSON{Raw: policyBytes},
						},
					},
				},
			},
		}
	}

	// network policy generation

	if polengine == "netpol" {
		generatedPolicyName := metadataMap["name"].(string)
		selector := specMap["podSelector"].(map[string]any)
		delete(selector, "matchLabels")
		selector["matchLabels"] = labels

		policyBytes, err := json.Marshal(policyData)

		if err != nil {
			return pol, err
		}
		pol = kyvernov1.Policy{
			ObjectMeta: metav1.ObjectMeta{
				Name: generatorPolicyName,
			},
			Spec: kyvernov1.Spec{
				GenerateExisting: true,
				Rules: []kyvernov1.Rule{
					{
						Name: cve + "virtual-patch-netpol",
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
						RawAnyAllConditions: &v1.JSON{Raw: preconditionBytes},
						Context: []kyvernov1.ContextEntry{
							getPodName,
						},
						Generation: kyvernov1.Generation{
							ResourceSpec: kyvernov1.ResourceSpec{
								APIVersion: "networking.k8s.io/v1",
								Kind:       "NetworkPolicy",
								Name:       generatedPolicyName,
								Namespace:  np.GetNamespace(),
							},
							RawData: &v1.JSON{Raw: policyBytes},
						},
					},
				},
			},
		}
	}

	return pol
}
