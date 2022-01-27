/*
Copyright 2022.

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SidecarSetSpec defines the desired state of SidecarSet
type SidecarSetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//query over pods that should be injected sidecar
	Selector   *metav1.LabelSelector `json:"selector,omitempty"`
	Containers []SidecarContainer    `json:"containers,omitempty"`
}

type SidecarContainer struct {
	v1.Container `json:",inline"`
}

// SidecarSetStatus defines the observed state of SidecarSet
type SidecarSetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//the number of Pods whose labels are matched with this SidecarSet's selector
	MatchedPods int32 `json:"matchedPods"`
	//the number of matched Pods that are injected with the latest SidecarSet's
	UpdatedPods int32 `json:"updatedPods"`
	//the number of matched Pods that have a ready condition
	ReadyPods int32 `json:"readyPods"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="MATCHED",type="integer",JSONPath=".status.matchedPods",description="The number of pods matched."
// +kubebuilder:printcolumn:name="UPDATED",type="integer",JSONPath=".status.updatedPods",description="The number of pods matched and updated."
// +kubebuilder:printcolumn:name="READY",type="integer",JSONPath=".status.readyPods",description="the number of matched Pods that have a ready condition."

// SidecarSet is the Schema for the sidecarsets API
type SidecarSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SidecarSetSpec   `json:"spec,omitempty"`
	Status SidecarSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SidecarSetList contains a list of SidecarSet
type SidecarSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SidecarSet `json:"items"`
}

func init() {
	//将Go类型添加到API组中
	SchemeBuilder.Register(&SidecarSet{}, &SidecarSetList{})
}
