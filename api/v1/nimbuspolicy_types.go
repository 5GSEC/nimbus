// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NimbusPolicySpec defines the desired state of NimbusPolicy
type NimbusPolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Selector specifies the target resources to which the policy applies
	Selector NimbusSelector `json:"selector"`

	// PolicyType specifies the type of policy, e.g., "Network", "System", "Cluster"
	NimbusRules []NimbusRules `json:"rules"`
}

// NimbusSelector is used to select specific resources based on labels.
type NimbusSelector struct {
	// MatchLabels is a map that holds key-value pairs to match against labels of resources.
	MatchLabels map[string]string `json:"matchLabels"`
}

// NimbusRules represents a single policy rule with an ID, type, description, and detailed rule configurations.
type NimbusRules struct {
	Id          string `json:"id"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Rule        []Rule `json:"rule"`
}

type Rule struct {
	RuleAction string `json:"action"`

	// Network: MatchProtocols
	MatchProtocols []MatchProtocol `json:"matchProtocols,omitempty"`

	// Process: MatchPaths, MatchDirectories, MatchPatterns
	// File: MatchPaths, MatchDirectories,  MatchPatterns
	MatchPaths       []MatchPath      `json:"matchPaths,omitempty"`
	MatchDirectories []MatchDirectory `json:"matchDirectories,omitempty"`
	MatchPatterns    []MatchPattern   `json:"matchPatterns,omitempty"`

	// Capabilities: MatchCapabilities
	MatchCapabilities []MatchCapability `json:"matchCapabilities,omitempty"`

	// Syscalls: MatchSyscalls
	MatchSyscalls     []MatchSyscall     `json:"matchSyscalls,omitempty"`
	MatchSyscallPaths []MatchSyscallPath `json:"matchSyscallPaths,omitempty"`

	FromCIDRSet []CIDRSet `json:"fromCIDRSet,omitempty"`
	ToPorts     []ToPort  `json:"toPorts,omitempty"`
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
	Directory  string             `json:"dir,omitempty"`
	FromSource []NimbusFromSource `json:"fromSource,omitempty"`
}

// MatchPattern defines a pattern for process policies
type MatchPattern struct {
	Pattern string `json:"pattern,omitempty"`
}

// MatchSyscall defines a syscall for syscall policies
type MatchSyscall struct {
	Syscalls   []string            `json:"syscalls,omitempty"`
	FromSource []SyscallFromSource `json:"fromSource,omitempty"`
}

type MatchSyscallPath struct {
	Path       string              `json:"path,omitempty"`
	Recursive  bool                `json:"recursive,omitempty"`
	Syscalls   []string            `json:"syscall,omitempty"`
	FromSource []SyscallFromSource `json:"fromSource,omitempty"`
}

type SyscallFromSource struct {
	Path string `json:"path,omitempty"`
	Dir  string `json:"dir,omitempty"`
}

// MatchCapability defines a capability for capabilities policies
type MatchCapability struct {
	Capability string             `json:"capability,omitempty"`
	FromSource []NimbusFromSource `json:"fromSource,omitempty"`
}

// FromSource defines a source path for directory-based policies
type NimbusFromSource struct {
	Path string `json:"path,omitempty"`
}

// NimbusPolicyStatus defines the observed state of NimbusPolicy
type NimbusPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	PolicyStatus string `json:"status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource: shortName="np"
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NimbusPolicy is the Schema for the nimbuspolicies API
type NimbusPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              NimbusPolicySpec   `json:"spec,omitempty"`
	Status            NimbusPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NimbusPolicyList contains a list of NimbusPolicy
type NimbusPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NimbusPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NimbusPolicy{}, &NimbusPolicyList{})
}
