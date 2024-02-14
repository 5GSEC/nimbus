// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NimbusPolicySpec defines the desired state of NimbusPolicy
type NimbusPolicySpec struct {
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
	ID          string `json:"id"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Rule        Rule   `json:"rule"`
}

type Rule struct {
	RuleAction string              `json:"action"`
	Params     map[string][]string `json:"params,omitempty"`
}

// NimbusPolicyStatus defines the observed state of NimbusPolicy
type NimbusPolicyStatus struct {
	Status      string      `json:"status"`
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
//+kubebuilder:resource: shortName="np"

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
