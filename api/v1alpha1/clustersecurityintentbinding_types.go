// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NamespaceSelector struct {
	MatchNames   []string `json:"matchNames,omitempty"`
	ExcludeNames []string `json:"excludeNames,omitempty"`
}

type ClusterMatchWorkloads struct {
	NodeSelector     LabelSelector     `json:"nodeSelector,omitempty"`
	NsSelector       NamespaceSelector `json:"nsSelector,omitempty"`
	WorkloadSelector LabelSelector     `json:"workloadSelector,omitempty"`
}

// ClusterSecurityIntentBindingSpec defines the desired state of ClusterSecurityIntentBinding
type ClusterSecurityIntentBindingSpec struct {
	Intents  []MatchIntent         `json:"intents"`
	Selector ClusterMatchWorkloads `json:"selector,omitempty"`
	CEL      []string              `json:"cel,omitempty"`
}

// ClusterSecurityIntentBindingStatus defines the observed state of ClusterSecurityIntentBinding
type ClusterSecurityIntentBindingStatus struct {
	Status               string      `json:"status"`
	LastUpdated          metav1.Time `json:"lastUpdated,omitempty"`
	NumberOfBoundIntents int32       `json:"numberOfBoundIntents"`
	BoundIntents         []string    `json:"boundIntents,omitempty"`
	ClusterNimbusPolicy  string      `json:"clusterNimbusPolicy"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,shortName="csib"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Intents",type="integer",JSONPath=".status.numberOfBoundIntents"
//+kubebuilder:printcolumn:name="ClusterNimbusPolicy",type="string",JSONPath=".status.clusterNimbusPolicy"
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
