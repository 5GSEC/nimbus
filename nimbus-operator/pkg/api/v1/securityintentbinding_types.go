// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecurityIntentBindingSpec defines the desired state of SecurityIntentBinding
type SecurityIntentBindingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of SecurityIntentBinding. Edit securityintentbinding_types.go to remove/update
	Intents  []MatchIntent `json:"intents"`
	Selector Selector      `json:"selector"`
}

type MatchIntent struct {
	Name string `json:"name"`
}

// Selector defines the selection criteria for resources
type Selector struct {
	Any []ResourceFilter `json:"any,omitempty"`
	All []ResourceFilter `json:"all,omitempty"`
	CEL []string         `json:"cel,omitempty"`
}

// ResourceFilter is used for filtering resources
type ResourceFilter struct {
	Resources Resources `json:"resources,omitempty"`
}

// Resources defines the properties for selecting Kubernetes resources
type Resources struct {
	Kind        string            `json:"kind,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

// SecurityIntentBindingStatus defines the observed state of SecurityIntentBinding
type SecurityIntentBindingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
// +kubebuilder:resource: shortName="sib"
//+kubebuilder:subresource:status

// SecurityIntentBinding is the Schema for the securityintentbindings API
type SecurityIntentBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SecurityIntentBindingSpec   `json:"spec,omitempty"`
	Status            SecurityIntentBindingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecurityIntentBindingList contains a list of SecurityIntentBinding
type SecurityIntentBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityIntentBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecurityIntentBinding{}, &SecurityIntentBindingList{})
}
