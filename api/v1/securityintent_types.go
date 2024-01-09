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
	// +kubebuilder:validation:Pattern:="^[a-zA-Z0-9]*$"
	Id          string                 `json:"id"`
	Description string                 `json:"description,omitempty"`
	Action      string                 `json:"action"`
	Mode        string                 `json:"mode"`
	Severity    int                    `json:"severity,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Params      []SecurityIntentParams `json:"params,omitempty"`
}

// Resource defines the resources that the security policy applies to
type SecurityIntentParams struct {
	// Network: MatchProtocols
	MatchProtocols []SecurityIntentMatchProtocol `json:"SecurityIntentMatchProtocols,omitempty"`

	// Process: MatchPaths, MatchDirectories, MatchPatterns
	// File: MatchPaths, MatchDirectories
	MatchPaths       []SecurityIntentMatchPath      `json:"matchPaths,omitempty"`
	MatchDirectories []SecurityIntentMatchDirectory `json:"matchDirectories,omitempty"`
	MatchPatterns    []SecurityIntentMatchPattern   `json:"matchPatterns,omitempty"`

	// Capabilities: MatchCapabilities
	MatchCapabilities []SecurityIntentMatchCapability `json:"matchCapabilities,omitempty"`

	// Syscalls: MatchSyscalls
	MatchSyscalls []SecurityIntentMatchSyscall `json:"matchSyscalls,omitempty"`

	FromCIDRSet []SecurityIntentCIDRSet `json:"fromCIDRSet,omitempty"`
	ToPorts     []SecurityIntentToPort  `json:"toPorts,omitempty"`
}

// CIDRSet defines CIDR ranges for network policies
type SecurityIntentCIDRSet struct {
	CIDR string `json:"cidr,omitempty"`
}

// ToPort defines ports and protocols for network policies
type SecurityIntentToPort struct {
	Ports []SecurityIntentPort `json:"ports,omitempty"`
}

// Port defines a network port and its protocol
type SecurityIntentPort struct {
	Port     string `json:"port,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

// SecurityIntentMatchProtocol defines a protocol for network policies
type SecurityIntentMatchProtocol struct {
	Protocol string `json:"protocol,omitempty"`
}

// MatchPath defines a path for process or file policies
type SecurityIntentMatchPath struct {
	Path string `json:"path,omitempty"`
}

// MatchDirectory defines a directory for process or file policies
type SecurityIntentMatchDirectory struct {
	Directory  string                     `json:"dir,omitempty"`
	FromSource []SecurityIntentFromSource `json:"fromSource,omitempty"`
}

// MatchPattern defines a pattern for process policies
type SecurityIntentMatchPattern struct {
	Pattern string `json:"pattern,omitempty"`
}

// MatchSyscall defines a syscall for syscall policies
type SecurityIntentMatchSyscall struct {
	Syscalls []string `json:"syscalls,omitempty"`
}

// MatchCapability defines a capability for capabilities policies
type SecurityIntentMatchCapability struct {
	Capability string `json:"capability,omitempty"`
}

type SecurityIntentFromSource struct {
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
// +kubebuilder:resource: shortName="si"
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
