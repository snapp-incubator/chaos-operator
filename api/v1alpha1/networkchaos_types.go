/*
Copyright 2024.

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

// NetworkChaosSpec defines the desired state of NetworkChaos
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NetworkChaosSpec struct {
	metav1.TypeMeta `json:",inline"`

	// +kubebuilder:validation:Required
	Upstream Upstream `json:"upstream"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=downstream
	Stream string `json:"stream"`

	// +kubebuilder:validation:Required
	LatencyToxic LatencyToxic `json:"latencyToxic"`
	TimeoutToxic TimeoutToxic `json:"timeoutToxic"`
}

// Upstream defines the upstream service details
type Upstream struct {

	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	Port string `json:"port"`
}

// Toxic defines the common structure of a toxic
type LatencyToxic struct {

	// +kubebuilder:validation:Required
	Latency int `json:"latency"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Maximum=1.0
	// +kubebuilder:validation:Minimum=0.0
	// +kubebuilder:default=1.0
	Probability float32 `json:"probability"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	Jitter int `json:"jitter"`
}
type TimeoutToxic struct {

	// +kubebuilder:validation:Required
	Timeout int `json:"timeout"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Maximum=1.0
	// +kubebuilder:validation:Minimum=0.0
	// +kubebuilder:default=1.0
	Probability float32 `json:"probability"`
}

// NetworkChaosStatus defines the observed state of NetworkChaos
type NetworkChaosStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NetworkChaos is the Schema for the networkchaos API
type NetworkChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkChaosSpec   `json:"spec,omitempty"`
	Status NetworkChaosStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NetworkChaosList contains a list of NetworkChaos
type NetworkChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkChaos `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkChaos{}, &NetworkChaosList{})
}
