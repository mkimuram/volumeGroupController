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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VolumeGroupSnapshotContentSpec defines the desired state of VolumeGroupSnapshotContent
type VolumeGroupSnapshotContentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of VolumeGroupSnapshotContent. Edit volumegroupsnapshotcontent_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// VolumeGroupSnapshotContentStatus defines the observed state of VolumeGroupSnapshotContent
type VolumeGroupSnapshotContentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VolumeGroupSnapshotContent is the Schema for the volumegroupsnapshotcontents API
type VolumeGroupSnapshotContent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeGroupSnapshotContentSpec   `json:"spec,omitempty"`
	Status VolumeGroupSnapshotContentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VolumeGroupSnapshotContentList contains a list of VolumeGroupSnapshotContent
type VolumeGroupSnapshotContentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeGroupSnapshotContent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolumeGroupSnapshotContent{}, &VolumeGroupSnapshotContentList{})
}
