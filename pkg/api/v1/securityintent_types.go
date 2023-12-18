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

// Intent defines the security policy details
type Intent struct {
	Description string     `json:"description,omitempty"` // Define the description
	Action      string     `json:"action"`                // Define the action of the policy
	Type        string     `json:"type"`                  // Defines the type of the policy
	Resource    []Resource `json:"resource"`              // Define the resources to which the security policy applies
}

// Resource defines the resources that the security policy applies to
type Resource struct {
	Network      []Network      `json:"network,omitempty"`
	Process      []Process      `json:"process,omitempty"`
	File         []File         `json:"file,omitempty"`
	Capabilities []Capabilities `json:"capabilities,omitempty"`
	Syscalls     []Syscalls     `json:"syscalls,omitempty"`
	FromCIDRSet  []CIDRSet      `json:"fromCIDRSet,omitempty"`
	ToPorts      []ToPort       `json:"toPorts,omitempty"`
}

// Network defines the network-related policies
type Network struct {
	MatchProtocols []MatchProtocol `json:"matchProtocols,omitempty"`
}

// Process defines the process-related policies
type Process struct {
	MatchPaths       []MatchPath      `json:"matchPaths,omitempty"`
	MatchDirectories []MatchDirectory `json:"matchDirectories,omitempty"`
	MatchPatterns    []MatchPattern   `json:"matchPatterns,omitempty"`
}

// File defines the file-related policies
type File struct {
	MatchPaths       []MatchPath      `json:"matchPaths,omitempty"`
	MatchDirectories []MatchDirectory `json:"matchDirectories,omitempty"`
}

// Capabilities defines the capabilities-related policies
type Capabilities struct {
	MatchCapabilities []MatchCapability `json:"matchCapabilities,omitempty"`
}

// Syscalls defines the syscalls-related policies
type Syscalls struct {
	MatchSyscalls []MatchSyscall `json:"matchSyscalls,omitempty"`
}

// CIDRSet defines CIDR ranges for network policies
type CIDRSet struct {
	CIDR string `json:"cidr,omitempty"`
}

// ToPort defines ports and protocols for network policies
type ToPort struct {
	Ports []Port `json:"ports,omitempty"`
}

// Port defines a network port and its protocol
type Port struct {
	Port     string `json:"port,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

// MatchProtocol defines a protocol for network policies
type MatchProtocol struct {
	Protocol string `json:"protocol,omitempty"`
}

// MatchPath defines a path for process or file policies
type MatchPath struct {
	Path string `json:"path,omitempty"`
}

// MatchDirectory defines a directory for process or file policies
type MatchDirectory struct {
	Directory  string       `json:"dir,omitempty"`
	FromSource []FromSource `json:"fromSource,omitempty"`
}

// MatchPattern defines a pattern for process policies
type MatchPattern struct {
	Pattern string `json:"pattern,omitempty"`
}

// MatchSyscall defines a syscall for syscall policies
type MatchSyscall struct {
	Syscalls []string `json:"syscalls,omitempty"`
}

// MatchCapability defines a capability for capabilities policies
type MatchCapability struct {
	Capability string `json:"capability,omitempty"`
}

// FromSource defines a source path for directory-based policies
type FromSource struct {
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
