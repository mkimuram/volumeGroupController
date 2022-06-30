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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeGroupSpec defines the desired state of VolumeGroup
type VolumeGroupSpec struct {
	// Selector is a label query over PersistentVolumeClaims that should match the volume group.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
}

//+kubebuilder:object:root=true

// VolumeGroup is the Schema for the volumegroups API
type VolumeGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VolumeGroupSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// VolumeGroupList contains a list of VolumeGroup
type VolumeGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolumeGroup{}, &VolumeGroupList{})
}
