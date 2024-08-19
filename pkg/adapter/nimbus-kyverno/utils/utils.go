// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetGVK(kind string) string {
	// Map to store the mappings of kinds to their corresponding API versions
	kindToAPIVersion := map[string]string{
		"deployment":            "apps/v1",
		"pod":                   "v1",
		"statefulset":           "apps/v1",
		"daemonset":             "apps/v1",
		"replicaset":            "apps/v1",
	}

	// Convert kind to lowercase to handle case insensitivity
	kind = strings.ToLower(kind)

	// Retrieve API version from the map
	apiVersion, exists := kindToAPIVersion[kind]
	if !exists {
		return "" 
	}

	switch kind {
	case "replicaset":
		kind = "ReplicaSet"
	case "statefulset":
		kind = "StatefulSet"
	case "daemonset" :
		kind = "DaemonSet"
	default:
		kind = Title(kind)
	}

	// Combine API version and kind to form the GroupVersionKind string
	return fmt.Sprintf("%s/%s", apiVersion, Title(kind))
}

// sort.Slice(planets, func(i, j int) bool {
// 	return planets[i].Axis < planets[j].Axis
//   })

func PolEqual(a, b kyvernov1.Policy) (string, bool) {
	if len(a.Spec.Rules[0].MatchResources.Any) != len(b.Spec.Rules[0].MatchResources.Any) {
		return "diff: labels not equal", false
	}
	if a.ObjectMeta.Name != b.ObjectMeta.Name {
		return "diff: name", false
	}
	if a.ObjectMeta.Namespace != b.ObjectMeta.Namespace {
		return "diff: Namespace", false
	}

	if !reflect.DeepEqual(a.ObjectMeta.Labels, b.ObjectMeta.Labels) {
		return "diff: Labels", false
	}

	if !reflect.DeepEqual(a.ObjectMeta.OwnerReferences, b.ObjectMeta.OwnerReferences) {
		return "diff: OwnerReferences", false
	}

	if !checkLabels(a, b) {
		return "diff: labels", false
	}

	if !reflect.DeepEqual(a.Spec, b.Spec) {
		return "diff: Spec", false
	}
	return "", true
}

func CheckIfReady(conditions []metav1.Condition) bool {
	for _, condition := range conditions {
		if condition.Type == "Ready" && condition.Reason == "Succeeded" {
			return true
		}
	}
	return false
}
func checkLabels(a, b kyvernov1.Policy) bool {
	resourceFiltersA := a.Spec.Rules[0].MatchResources.Any
	resourceFiltersB := b.Spec.Rules[0].MatchResources.Any
	if len(resourceFiltersA) != len(resourceFiltersB) {
		return false
	}
	mp := make(map[string]bool)
	for _, filter := range resourceFiltersA {
		if filter.Selector != nil {
			for k,v := range filter.Selector.MatchLabels {
				key := k+v
				mp[key] = true
			}
		}
	}

	for _, filter := range resourceFiltersB {
		if filter.Selector != nil {
			for k,v := range filter.Selector.MatchLabels {
				key := k+v
				if !mp[key] {
					return false
				}
			}
		}
	}
	return true
}
func Title(input string) string {
    toTitle := cases.Title(language.Und)

    return toTitle.String(input)
}

func GetData[T any]()(T, error) {
	var out T
	// Open the JSON file
	file, err := os.Open("../../../vp.json")
	if err != nil {
		return out, err
		// fmt.Println(err)
	}
	defer file.Close()

	// Read the file contents
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return out, err
		// fmt.Println(err)
	}

	err = json.Unmarshal(bytes, &out)
	if err != nil {
		return out, err
		// fmt.Println(err)
	}

	return out, nil
}

func Contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}