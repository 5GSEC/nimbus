// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterNimbusPolicySpec defines the desired state of ClusterNimbusPolicy
type ClusterNimbusPolicySpec struct {
	Selector    CwSelector    `json:"selector"`
	NimbusRules []NimbusRules `json:"rules"`
}

// ClusterNimbusPolicyStatus defines the observed state of ClusterNimbusPolicy
type ClusterNimbusPolicyStatus struct {
	Status string `json:"status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,shortName="cwnp"
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterNimbusPolicy is the Schema for the clusternimbuspolicies API
type ClusterNimbusPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterNimbusPolicySpec   `json:"spec,omitempty"`
	Status ClusterNimbusPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterNimbusPolicyList contains a list of ClusterNimbusPolicy
type ClusterNimbusPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterNimbusPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterNimbusPolicy{}, &ClusterNimbusPolicyList{})
}
