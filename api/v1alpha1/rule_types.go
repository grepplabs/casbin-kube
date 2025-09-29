package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RuleSpec defines the desired state of Rule.
type RuleSpec struct {
	// Rule type: p, p2, g, g2, ...
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^(p|g)\d*$`
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="ptype is immutable"
	PType string `json:"ptype"`

	// Positional parameters v0
	// +kubebuilder:selectablefield:JSONPath=.spec.v0
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="v0 is immutable"
	V0 string `json:"v0,omitempty"`

	// Positional parameters v1
	// +kubebuilder:validation:Pattern=`.*\S.*`
	// +kubebuilder:selectablefield:JSONPath=.spec.v1
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="v1 is immutable"
	V1 string `json:"v1,omitempty"`

	// Positional parameters v2
	// +kubebuilder:selectablefield:JSONPath=.spec.v2
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="v2 is immutable"
	V2 string `json:"v2,omitempty"`

	// Positional parameters v3
	// +kubebuilder:selectablefield:JSONPath=.spec.v3
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="v3 is immutable"
	V3 string `json:"v3,omitempty"`

	// Positional parameters v4
	// +kubebuilder:selectablefield:JSONPath=.spec.v4
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="v4 is immutable"
	V4 string `json:"v4,omitempty"`

	// Positional parameters v5
	// +kubebuilder:selectablefield:JSONPath=.spec.v5
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="v5 is immutable"
	V5 string `json:"v5,omitempty"`
}

// RuleStatus defines the observed state of Rule.
type RuleStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="PType",type="string",JSONPath=`.spec.ptype`
// +kubebuilder:printcolumn:name="V0",type="string",JSONPath=`.spec.v0`
// +kubebuilder:printcolumn:name="V1",type="string",JSONPath=`.spec.v1`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:selectablefield:JSONPath=.spec.ptype
// +kubebuilder:selectablefield:JSONPath=.spec.v0
// +kubebuilder:selectablefield:JSONPath=.spec.v1
// +kubebuilder:selectablefield:JSONPath=.spec.v2
// +kubebuilder:selectablefield:JSONPath=.spec.v3
// +kubebuilder:selectablefield:JSONPath=.spec.v4
// +kubebuilder:selectablefield:JSONPath=.spec.v5

// Rule is the Schema for the policies API.
type Rule struct {
	metav1.TypeMeta `json:",inline"`
	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`
	// spec defines the desired state of Rule
	// +required
	Spec RuleSpec `json:"spec"`
	// status defines the observed state of Rule
	// +optional
	Status RuleStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// RuleList contains a list of Rule.
type RuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Rule{}, &RuleList{})
}
