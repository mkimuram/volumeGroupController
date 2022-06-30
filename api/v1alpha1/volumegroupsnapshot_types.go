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

// VolumeGroupSnapshotSpec defines the desired state of VolumeGroupSnapshot
type VolumeGroupSnapshotSpec struct {
	// +optional
	VolumeGroupName *string `json:"volumeGroupName,omitempty"`

	// +optional
	BoundVolumeGroupSnapshotContentName *string `json:"boundVolumeGroupSnapshotContentName,omitempty"`
}

// VolumeGroupSnapshotStatus defines the observed state of VolumeGroupSnapshot
type VolumeGroupSnapshotStatus struct {
	// ReadyToUse becomes true when ReadyToUse on all individual snapshots become true
	// +optional
	ReadyToUse *bool `json:"readyToUse,omitempty"`

	// +optional
	CreationTime *metav1.Time `json:"creationTime,omitempty"`

	// +optional
	Error *VolumeGroupSnapshotError `json:"error,omitempty"`
}

// VolumeGroupSnapshotError describes an error encountered on the group snapshot
type VolumeGroupSnapshotError struct {
	// time is the timestamp when the error was encountered.
	// +optional
	Time *metav1.Time `json:"time,omitempty"`

	// message details the encountered error
	// +optional
	Message *string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced,shortName=vgs
//+kubebuilder:printcolumn:name="ReadyToUse",type=boolean,JSONPath=`.status.readyToUse`,description="Indicates if the volumeGroupSnapshot is ready to be used to restore a volume."
//+kubebuilder:printcolumn:name="VolumeGroup",type=string,JSONPath=`.spec.volumeGroupName`,description="If a new volumeGroupSnapshotContent needs to be created, this contains the name of the volumeGroupName from which this volumeGroupSnapshot was (or will be) created."
//+kubebuilder:printcolumn:name="VolumeGroupSnapshotContent",type=string,JSONPath=`.spec.boundVolumeGroupSnapshotContentName`,description="Name of the VolumeGroupSnapshotContent object to which the VolumeGroupSnapshot object intends to bind to."

// VolumeGroupSnapshot is the Schema for the volumegroupsnapshots API
type VolumeGroupSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeGroupSnapshotSpec   `json:"spec,omitempty"`
	Status VolumeGroupSnapshotStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VolumeGroupSnapshotList contains a list of VolumeGroupSnapshot
type VolumeGroupSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeGroupSnapshot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolumeGroupSnapshot{}, &VolumeGroupSnapshotList{})
}
