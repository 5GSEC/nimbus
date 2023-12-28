// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

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

	Intent Intent `json:"intent"` // Define the details of the security policy.
}

// Intent defines a security intention that can be used to generate multiple security policies.
type Intent struct {
	// ID is predefined in Security Intent pool. It uniquely identifies a specific security intent.
	ID string `json:"id"`

	// Description is human-readable explanation of the intent's purpose.
	Description string `json:"description,omitempty"`

	// Action defines how the security policy will be enforced.
	Action string `json:"action"`

	// Mode defines the enforcement behavior of the intent.
	// Defaults to best-effort.
	Mode string `json:"mode,omitempty"`

	// Severity defines the potential impact of a security violation related to the intent.
	// Defaults to Low.
	Severity string `json:"severity,omitempty"`

	// Tags are additional metadata for categorization and grouping of intents.
	// Facilitates searching, filtering, and management of security policies.
	Tags []string `json:"tags,omitempty"`

	// Params are key-value pairs that allows fine-tuning of intents to specific requirements.
	Params map[string][]string `json:"params,omitempty"`
}

// SecurityIntentStatus defines the observed state of SecurityIntent
type SecurityIntentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// This field can be updated to reflect the actual status of the application of the security intents
}

// SecurityIntent is the Schema for the securityintents API
// +kubebuilder:object:root=true
// +kubebuilder:resource: shortName="sit"
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
