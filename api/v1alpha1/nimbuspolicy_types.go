// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NimbusPolicySpec defines the desired state of NimbusPolicy
type NimbusPolicySpec struct {
	// Selector specifies the target resources to which the policy applies
	Selector WorkloadSelector `json:"selector"`

	// PolicyType specifies the type of policy, e.g., "Network", "System", "Cluster"
	NimbusRules []NimbusRules `json:"rules"`
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
	Status                  string      `json:"status"`
	LastUpdated             metav1.Time `json:"lastUpdated,omitempty"`
	NumberOfAdapterPolicies int32       `json:"numberOfAdapterPolicies"`
	Policies                []string    `json:"adapterPolicies,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource: shortName="np"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Policies",type="integer",JSONPath=".status.numberOfAdapterPolicies"

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

// Check equality of the spec to decide if we need to update the object
func (a NimbusPolicy) Equal(b NimbusPolicy) (string, bool) {
	if a.ObjectMeta.Name != b.ObjectMeta.Name {
		return "diff: name", false
	}
	if a.ObjectMeta.Namespace != b.ObjectMeta.Namespace {
		return "diff: Namespace", false
	}

	if !reflect.DeepEqual(a.ObjectMeta.Labels, b.ObjectMeta.Labels) {
		return "diff: Labels", false
	}

	if !reflect.DeepEqual(a.ObjectMeta.OwnerReferences, b.ObjectMeta.OwnerReferences) {
		return "diff: OwnerReferences", false
	}

	if !reflect.DeepEqual(a.Spec, b.Spec) {
		return "diff: Spec", false
	}
	return "", true
}
