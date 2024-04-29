// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package v1alpha

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecurityIntentSpec defines the desired state of SecurityIntent
type SecurityIntentSpec struct {
	Intent Intent `json:"intent"` // Define the details of the security policy.
}

// Intent defines the security policy details
type Intent struct {
	// ID is predefined in adapter ID pool.
	// Used by security engines to generate corresponding security policies.
	//+kubebuilder:validation:Pattern:="^[a-zA-Z0-9]*$"
	ID string `json:"id"`

	// Description is human-readable explanation of the intent's purpose.
	Description string `json:"description,omitempty"`

	// Action defines how the security policy will be enforced.
	Action string `json:"action"`

	// Severity defines the potential impact of a security violation related to the intent.
	// Defaults to Low.
	//+kubebuilder:default:=Low
	Severity string `json:"severity,omitempty"`

	// Tags are additional metadata for categorization and grouping of intents.
	// Facilitates searching, filtering, and management of security policies.
	Tags []string `json:"tags,omitempty"`

	// Params are key-value pairs that allows fine-tuning of intents to specific requirements.
	Params map[string][]string `json:"params,omitempty"`
}

// SecurityIntentStatus defines the observed state of SecurityIntent
type SecurityIntentStatus struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	Status string `json:"status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName="si",scope="Cluster"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".spec.intent.id",priority=1
// +kubebuilder:printcolumn:name="Action",type="string",JSONPath=".spec.intent.action",priority=1
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
