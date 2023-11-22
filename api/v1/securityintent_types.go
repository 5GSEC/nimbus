/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecurityIntentSpec defines the desired state of SecurityIntent
type SecurityIntentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Selector Selector `json:"selector"` // Define criteria for selecting resources
	Intent   Intent   `json:"intent"`   // Define the details of the security policy.
}

// Selector defines the selection criteria for resources
type Selector struct {
	Match Match    `json:"match,omitempty"` // Define the resource filter to be used
	CEL   []string `json:"cel"`             // Define filter conditions as CEL expressions
}

// Match defines the resource filters to be used
type Match struct {
	Any []ResourceFilter `json:"any,omitempty"` // Apply when one or more conditions match
	All []ResourceFilter `json:"all,omitempty"` //Apply when all conditions must match
}

// ResourceFilter is used for filtering resources, subjects, roles, and cluster roles
type ResourceFilter struct {
	Resources Resources `json:"resources,omitempty"` // Define properties to select k8s resources
	Subjects  []Subject `json:"subjects,omitempty"`  // Define the subjects to filter
	Roles     []string  `json:"roles,omitempty"`     // Define the roles to filter.
}

// Resources defines the properties for selecting Kubernetes resources
type Resources struct {
	Names      []string `json:"names,omitempty"`      // Define the resource name
	Namespaces []string `json:"namespaces,omitempty"` // Define the namespaces to which the resource belongs
	Kinds      []string `json:"kinds"`                // Define resource kinds
	Operations []string `json:"operations,omitempty"` // Define operations for the resource

	MatchLabels map[string]string `json:"matchLabels,omitempty"` // Define labels to apply to the resource
}

// Subject defines the subject for filtering
type Subject struct {
	Kind string `json:"kind"`           // Define the kind of policy
	Name string `json:"name,omitempty"` // Define the name of the policy
}

// Intent defines the security policy details
type Intent struct {
	Action   string     `json:"action"`   // Define the action of the policy
	Mode     string     `json:"mode"`     // Defines the mode of the policy
	Type     string     `json:"type"`     // Defines the type of the policy
	Resource []Resource `json:"resource"` // Define the resource to which the security policy applies
}

// Resource defines the resources that the security policy applies to
type Resource struct {
	Key    string   `json:"key,omitempty"`    // Define a resource key
	Val    []string `json:"val,omitempty"`    // Define a resource value list
	Valcel string   `json:"valcel,omitempty"` // Define a CEL expression
	Attrs  []string `json:"attrs,omitempty"`  // Define additional attributes
}

// SecurityIntentStatus defines the observed state of SecurityIntent
type SecurityIntentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// This field can be updated to reflect the actual status of the application of the security intents
}

// SecurityIntent is the Schema for the securityintents API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SecurityIntent is the Schema for the securityintents API
type SecurityIntent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SecurityIntentSpec   `json:"spec,omitempty"`
	Status            SecurityIntentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecurityIntentList contains a list of SecurityIntent
type SecurityIntentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityIntent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecurityIntent{}, &SecurityIntentList{})
}
