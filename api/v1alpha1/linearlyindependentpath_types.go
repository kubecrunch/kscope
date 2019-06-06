/*
Copyright 2019 cskkman@gmail.com.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type KscopePreservedFields struct {
	FieldName string `json:"field_name"`
	ValueType string `json:"value_type"`
	KeyName   string `json:"key_name"`
}

// Duct stores information captured at various stages within a linearly Independent path so that it can be reused
// ex. the first stage ie. sequence_nunber with 1 acquires jwt token which is used in later stages to make furhter api calls.
//type Duct map[string]KscopePreservedFields

// KscopeRequest is the http request
type KscopeRequest struct {
	Method  string            `json:"method"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
	Url     string            `json:"url"`
}

// KscopeResponse
type KscopeResponse struct {
	StatusCode            int                     `json:"status_code"`
	MaxPermissibleLatency int                     `json:"max_permissible_latency"` // in milli seconds
	ExpectedFields        []string                `json:"expected_fields"`         // in later versions .. regex checks can be made on values
	PreserveFields        []KscopePreservedFields `json:"preserve_fields"`
}

type KscopeStage struct {
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	SequenceNumber int            `json:"sequence_number"`
	Request        KscopeRequest  `json:"request"`
	Response       KscopeResponse `json:"response"`
}

// LinearlyIndependentPathSpec defines the desired state of LinearlyIndependentPath
type LinearlyIndependentPathSpec struct {
	Stages              []KscopeStage `json:"stages"`
	BootStrappedSecrets []string      `json:"bootstrapped_secrets"`
	// Important: Run "make" to regenerate code after modifying this file
}

// LinearlyIndependentPathStatus defines the observed state of LinearlyIndependentPath
type LinearlyIndependentPathStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// LinearlyIndependentPath is the Schema for the linearlyindependentpaths API
type LinearlyIndependentPath struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LinearlyIndependentPathSpec   `json:"spec,omitempty"`
	Status LinearlyIndependentPathStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LinearlyIndependentPathList contains a list of LinearlyIndependentPath
type LinearlyIndependentPathList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LinearlyIndependentPath `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LinearlyIndependentPath{}, &LinearlyIndependentPathList{})
}
