// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CwResource struct {
	Kind        string            `json:"kind"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace,omitempty"`
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type CwSelector struct {
	Resources []CwResource `json:"resources,omitempty"`
	CEL       []string     `json:"cel,omitempty"`
}

// ClusterSecurityIntentBindingSpec defines the desired state of ClusterSecurityIntentBinding
type ClusterSecurityIntentBindingSpec struct {
	Intents  []MatchIntent `json:"intents"`
	Selector CwSelector    `json:"selector"`
}

// ClusterSecurityIntentBindingStatus defines the observed state of ClusterSecurityIntentBinding
type ClusterSecurityIntentBindingStatus struct {
	Status string `json:"status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
//+kubebuilder:resource:scope=Cluster,shortName="csib"
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterSecurityIntentBinding is the Schema for the clustersecurityintentbindings API
type ClusterSecurityIntentBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSecurityIntentBindingSpec   `json:"spec,omitempty"`
	Status ClusterSecurityIntentBindingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterSecurityIntentBindingList contains a list of ClusterSecurityIntentBinding
type ClusterSecurityIntentBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterSecurityIntentBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterSecurityIntentBinding{}, &ClusterSecurityIntentBindingList{})
}
