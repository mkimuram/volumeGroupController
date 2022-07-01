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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	volumegroupv1alpha1 "github.com/mkimuram/volumeGroupController/api/v1alpha1"
)

// VolumeGroupSnapshotReconciler reconciles a VolumeGroupSnapshot object
type VolumeGroupSnapshotReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=volumegroup.example.com,resources=volumegroupsnapshots,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=volumegroup.example.com,resources=volumegroupsnapshots/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=volumegroup.example.com,resources=volumegroupsnapshots/finalizers,verbs=update
//+kubebuilder:rbac:groups=volumegroup.example.com,resources=volumegroups,verbs=get
//+kubebuilder:rbac:groups=volumegroup.example.com,resources=volumegroupsnapshotContents,verbs=get;create
//+kubebuilder:rbac:groups=volumegroup.example.com,resources=volumegroupsnapshotContents/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;

// Reconcile is reconciliation loop for VolumeGroupSnapshot
func (r *VolumeGroupSnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	vgs := &volumegroupv1alpha1.VolumeGroupSnapshot{}
	if err := r.Get(ctx, req.NamespacedName, vgs); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found. Ignore this
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if vgs.Status.ReadyToUse != nil && *vgs.Status.ReadyToUse {
		// Already ready to use
		return ctrl.Result{}, nil
	}

	if vgs.Spec.BoundVolumeGroupSnapshotContentName == nil {
		if vgs.Spec.VolumeGroupName != nil {
			// Create VolumeGroupSnapshotContent for VolumeGroup
			err := r.createVolumeGroupSnapshotContent(ctx, vgs)
			if err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{Requeue: true}, nil
		}
		// Retry until BoundVolumeGroupSnapshotContentName become non-nil.
		return ctrl.Result{Requeue: true}, nil
	}

	// Update ReadyToUse
	readyToUse, err := r.updateReadyToUse(ctx, vgs)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !readyToUse {
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *VolumeGroupSnapshotReconciler) createVolumeGroupSnapshotContent(ctx context.Context, vgs *volumegroupv1alpha1.VolumeGroupSnapshot) error {
	vgsc, err := r.volumeGroupSnapshotContentFor(ctx, vgs)
	if err != nil {
		return err
	}

	err = r.Create(ctx, vgsc)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
		// Continue already exists case
	}

	// Set vgsc.Name to vgs's VolumeGroupSnapshotContentName
	vgs.Spec.BoundVolumeGroupSnapshotContentName = &vgsc.Name
	// TODO: Consider also setting CreationTime somewhere

	if err := r.Update(ctx, vgs); err != nil {
		return err
	}

	return nil
}

func (r *VolumeGroupSnapshotReconciler) volumeGroupSnapshotContentFor(ctx context.Context, vgs *volumegroupv1alpha1.VolumeGroupSnapshot) (*volumegroupv1alpha1.VolumeGroupSnapshotContent, error) {
	if vgs.Spec.VolumeGroupName == nil {
		return nil, fmt.Errorf("VolumeGroupName for %s/%s is nill", vgs.Namespace, vgs.Name)
	}

	vg := &volumegroupv1alpha1.VolumeGroup{}

	if err := r.Get(ctx, types.NamespacedName{Name: *vgs.Spec.VolumeGroupName, Namespace: vgs.Namespace}, vg); err != nil {
		return nil, err
	}

	pvcList := &corev1.PersistentVolumeClaimList{}
	selector, err := metav1.LabelSelectorAsSelector(vg.Spec.Selector)
	if err != nil {
		return nil, err
	}
	//listOpts := []client.ListOption{labels.Set(selector.String()).String()}
	listOpts := &client.ListOptions{LabelSelector: selector}

	if err := r.List(ctx, pvcList, listOpts); err != nil {
		return nil, err
	}

	vgsc := &volumegroupv1alpha1.VolumeGroupSnapshotContent{
		ObjectMeta: metav1.ObjectMeta{
			// TODO: Consider generating a better name for VolumeGroupSnapshotContent from vgs.Name
			Name:      fmt.Sprintf("vgsc-%s", vgs.Name),
			Namespace: vgs.Namespace,
		},
		Spec: volumegroupv1alpha1.VolumeGroupSnapshotContentSpec{
			VolumeGroupSnapshotName:   &vgs.Name,
			PersistentVolumeClaimList: []string{},
			SnapshotList:              []string{},
		},
	}

	// Set all PVC's names to PersistentVolumeClaimList
	for _, pvc := range pvcList.Items {
		vgsc.Spec.PersistentVolumeClaimList = append(vgsc.Spec.PersistentVolumeClaimList, pvc.Name)
	}

	// Set owner reference from vgs to vgsc
	ctrl.SetControllerReference(vgs, vgsc, r.Scheme)

	return vgsc, nil
}

func (r *VolumeGroupSnapshotReconciler) updateReadyToUse(ctx context.Context, vgs *volumegroupv1alpha1.VolumeGroupSnapshot) (bool, error) {
	if vgs.Spec.BoundVolumeGroupSnapshotContentName == nil {
		return false, fmt.Errorf("BoundVolumeGroupSnapshotContentName for %s/%s is nill", vgs.Namespace, vgs.Name)
	}

	vgsc := &volumegroupv1alpha1.VolumeGroupSnapshotContent{}

	if err := r.Get(ctx, types.NamespacedName{Name: *vgs.Spec.BoundVolumeGroupSnapshotContentName, Namespace: vgs.Namespace}, vgsc); err != nil {
		return false, err
	}

	if vgsc.Status.ReadyToUse == nil || !*vgsc.Status.ReadyToUse {
		// VolumeGroupSnapshotContent for this VolumeGroupSnapshot isn't ready to use yet
		return false, nil
	}

	// Update VolumeGroupSnapshot's ReadyToUse to true
	vgs.Status.ReadyToUse = vgsc.Status.ReadyToUse

	if err := r.Status().Update(ctx, vgs); err != nil {
		return false, err
	}

	return true, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolumeGroupSnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&volumegroupv1alpha1.VolumeGroupSnapshot{}).
		Complete(r)
}
