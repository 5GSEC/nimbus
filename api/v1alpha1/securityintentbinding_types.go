// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecurityIntentBindingSpec defines the desired state of SecurityIntentBinding
type SecurityIntentBindingSpec struct {
	Intents  []MatchIntent  `json:"intents"`
	Selector MatchWorkloads `json:"selector"`
	CEL      []string       `json:"cel,omitempty"`
}

// MatchIntent struct defines the request for a specific SecurityIntent
type MatchIntent struct {
	Name string `json:"name"`
}

// Selector defines the selection criteria for resources
type MatchWorkloads struct {
	WorkloadSelector LabelSelector `json:"WorkloadSelector,omitempty"`
}

type LabelSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

// SecurityIntentBindingStatus defines the observed state of SecurityIntentBinding
type SecurityIntentBindingStatus struct {
	Status               string      `json:"status"`
	LastUpdated          metav1.Time `json:"lastUpdated,omitempty"`
	NumberOfBoundIntents int32       `json:"numberOfBoundIntents"`
	BoundIntents         []string    `json:"boundIntents,omitempty"`
	NimbusPolicy         string      `json:"nimbusPolicy"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource: shortName="sib"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Intents",type="integer",JSONPath=".status.numberOfBoundIntents"
// +kubebuilder:printcolumn:name="NimbusPolicy",type="string",JSONPath=".status.nimbusPolicy"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

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
