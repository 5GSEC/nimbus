// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecurityIntentSpec defines the desired state of SecurityIntent
type SecurityIntentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Intents []Intent `json:"intent"` // Define the details of the security policy.
}

// Intent defines the security policy details
type Intent struct {
	Description string       `json:"description,omitempty"` // Define the description
	Group       string       `json:"type"`                  // Defines the type of the policy
	ID          string       `json:"resource"`              // Define the resources to which the security policy applies
	Params      IntentParams `json:"params"`
}

// Resource defines the resources that the security policy applies to
type IntentParams struct {
	File ProtectFile `json:"protectFile,omitempty"`
	Port ProtectPort `json:"protectPort,omitempty"`

	// Only Owner can access file
	OwnerOnly File `json:"ownerOnly,omitempty"`

	// File cannot be accessed by anybody
	BlockAsset File

	// BlockRawSocket: does not have parameters
}

// ProtectFile will ensure only AllowBinaries can access the File
type ProtectFile struct {
	File          string `json:"port,omitempty"`
	AllowBinaries File   `json:"allowBinaries,omitempty"`
}

// ProtectPort will ensure only AllowBinaries can access Port
type ProtectPort struct {
	Port          string `json:"port,omitempty"`
	AllowBinaries File   `json:"allowBinaries,omitempty"`
}

// File defines the file-related policies
type File struct {
	MatchPaths []MatchPath `json:"matchPaths,omitempty"`
}

// MatchPath defines a path for process or file policies
type MatchPath struct {
	Path string `json:"path,omitempty"`
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
