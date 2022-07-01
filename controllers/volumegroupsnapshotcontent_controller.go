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

package controllers

import (
	"context"
	"fmt"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	volumegroupv1alpha1 "github.com/mkimuram/volumeGroupController/api/v1alpha1"
)

// VolumeGroupSnapshotContentReconciler reconciles a VolumeGroupSnapshotContent object
type VolumeGroupSnapshotContentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=volumegroup.example.com,resources=volumegroupsnapshotcontents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=volumegroup.example.com,resources=volumegroupsnapshotcontents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=volumegroup.example.com,resources=volumegroupsnapshotcontents/finalizers,verbs=update
//+kubebuilder:rbac:groups=snapshot.storage.k8s.io,resources=volumesnapshots,verbs=get;create

// Reconcile is reconciliation loop for VolumeGroupSnapshotContent
func (r *VolumeGroupSnapshotContentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	vgsc := &volumegroupv1alpha1.VolumeGroupSnapshotContent{}
	if err := r.Get(ctx, req.NamespacedName, vgsc); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found. Ignore this
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if vgsc.Status.ReadyToUse != nil && *vgsc.Status.ReadyToUse {
		// Already ready to use
		return ctrl.Result{}, nil
	}

	pvcs, err := r.getSnapshotMissingVolumes(ctx, vgsc)
	if err != nil {
		return ctrl.Result{}, err
	}

	if len(pvcs) > 0 {
		// Create VolumeSnapshot for pvcs
		err := r.createVolumeSnapshots(ctx, vgsc, pvcs)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	// Update ReadyToUse
	readyToUse, err := r.updateReadyToUse(ctx, vgsc)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !readyToUse {
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *VolumeGroupSnapshotContentReconciler) getSnapshotMissingVolumes(ctx context.Context, vgsc *volumegroupv1alpha1.VolumeGroupSnapshotContent) ([]string, error) {
	requiredPvcs := make(map[string]bool, len(vgsc.Spec.PersistentVolumeClaimList))

	// Set all pvcs in PersistentVolumeClaimList to requiredPvcs
	for _, pvc := range vgsc.Spec.PersistentVolumeClaimList {
		requiredPvcs[pvc] = true
	}

	// Check if a snapshot for each PVC exists in SnapshotList
	for _, vsName := range vgsc.Spec.SnapshotList {
		vs := &snapshotv1.VolumeSnapshot{}

		err := r.Get(ctx, types.NamespacedName{Name: vsName, Namespace: vgsc.Namespace}, vs)
		if err != nil {
			return []string{}, err
		}

		if vs.Spec.Source.PersistentVolumeClaimName == nil {
			continue
		}

		// Delete PersistentVolumeClaimName from requiredPvcs if exists
		if _, ok := requiredPvcs[*vs.Spec.Source.PersistentVolumeClaimName]; ok {
			delete(requiredPvcs, *vs.Spec.Source.PersistentVolumeClaimName)
		}
	}

	// Covert map to array
	volumes := []string{}
	for pvc := range requiredPvcs {
		volumes = append(volumes, pvc)
	}

	return volumes, nil
}

func (r *VolumeGroupSnapshotContentReconciler) createVolumeSnapshots(ctx context.Context, vgsc *volumegroupv1alpha1.VolumeGroupSnapshotContent, pvcs []string) error {
	for _, pvcName := range pvcs {
		vs := r.volumeSnapshotFor(ctx, vgsc, pvcName)

		if err := r.Create(ctx, vs); err != nil {
			if !errors.IsAlreadyExists(err) {
				return err
			}
			// Continue already exists case
		}

		// Add vs.Name to VolumeGroupSnapshotContent's SnapshotList
		vgsc.Spec.SnapshotList = append(vgsc.Spec.SnapshotList, vs.Name)
		// TODO: Consider also setting CreationTime somewhere

		if err := r.Update(ctx, vgsc); err != nil {
			return err
		}

	}
	return nil
}

func (r *VolumeGroupSnapshotContentReconciler) volumeSnapshotFor(ctx context.Context, vgsc *volumegroupv1alpha1.VolumeGroupSnapshotContent, pvcName string) *snapshotv1.VolumeSnapshot {
	vs := &snapshotv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			// TODO: Consider generating a better name for VolumeSnapshot from vgsc.Name and pvcName
			Name:      fmt.Sprintf("vs-%s-%s", vgsc.Name, pvcName),
			Namespace: vgsc.Namespace,
		},
		Spec: snapshotv1.VolumeSnapshotSpec{
			Source: snapshotv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvcName,
			},
			// TODO: VolumeSnapshotClassName should be passed so that non-default VolumeSnapshotClass can be used
			// Should it be per PVC? If not, VolumeSnapshotClassName can be added to VolumeGroupSnapshot's spec and VolumeGroupSnapshotContent's spec directly
		},
	}

	ctrl.SetControllerReference(vgsc, vs, r.Scheme)

	return vs
}

func (r *VolumeGroupSnapshotContentReconciler) updateReadyToUse(ctx context.Context, vgsc *volumegroupv1alpha1.VolumeGroupSnapshotContent) (bool, error) {
	for _, vsName := range vgsc.Spec.SnapshotList {
		vs := &snapshotv1.VolumeSnapshot{}

		if err := r.Get(ctx, types.NamespacedName{Name: vsName, Namespace: vgsc.Namespace}, vs); err != nil {
			return false, err
		}

		if vs.Status == nil || vs.Status.ReadyToUse == nil || !*vs.Status.ReadyToUse {
			// This VolumeSnapshot isn't ready to use
			return false, nil
		}
	}

	// Update VolumeGroupSnapshotContent's ReadyToUse to true
	ready := true
	vgsc.Status.ReadyToUse = &ready

	if err := r.Status().Update(ctx, vgsc); err != nil {
		return false, err
	}

	// All VolumeSnapshots in vgsc.Spec.SnapshotList are ready to use
	return true, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolumeGroupSnapshotContentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&volumegroupv1alpha1.VolumeGroupSnapshotContent{}).
		Complete(r)
}
