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

// VolumeGroupSnapshotContentSpec defines the desired state of VolumeGroupSnapshotContent
type VolumeGroupSnapshotContentSpec struct {
	// Required
	// VolumeGroupSnapshotRef specifies the VolumeGroupSnapshot object
	// to which this VolumeGroupSnapshotContent object is bound.
	VolumeGroupSnapshotName *string `json:"volumeGroupSnapshotName,omitempty"`

	// List of persistent volume claims to take snapshots from
	// +optional
	PersistentVolumeClaimList []string `json:"persistentVolumeClaimList"`

	// Required
	// List of volume snapshots
	SnapshotList []string `json:"snapshotList"`
}

// VolumeGroupSnapshotContentStatus defines the observed state of VolumeGroupSnapshotContent
type VolumeGroupSnapshotContentStatus struct {
	// ReadyToUse becomes true when ReadyToUse on all individual snapshots become true
	// +optional
	ReadyToUse *bool `json:"readyToUse,omitempty"`

	// +optional
	CreationTime *int64 `json:"creationTime,omitempty"`

	// +optional
	Error *VolumeGroupSnapshotError `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced,shortName=vgsc
//+kubebuilder:printcolumn:name="ReadyToUse",type=boolean,JSONPath=`.status.readyToUse`,description="Indicates if the volumeGroupSnapshotContent is ready to be used to restore a volume."
//+kubebuilder:printcolumn:name="VolumeGroupSnapshot",type=string,JSONPath=`.spec.volumeGroupSnapshotName`,description="Name of the VolumeGroupSnapshot object to which this VolumeGroupSnapshotContent object is bound."

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
