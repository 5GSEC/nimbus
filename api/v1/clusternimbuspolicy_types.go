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
	Status                  string      `json:"status"`
	LastUpdated             metav1.Time `json:"lastUpdated,omitempty"`
	NumberOfAdapterPolicies int32       `json:"numberOfAdapterPolicies"`
	Policies                []string    `json:"adapterPolicies,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,shortName="cwnp"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Policies",type="integer",JSONPath=".status.numberOfAdapterPolicies"

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
