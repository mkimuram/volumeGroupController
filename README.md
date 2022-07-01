# volumegroupcontroller
PoC implementation of "Add Volume Group" KEP(#1551)

## Description
This implementation uses label selector approach. Also, `VolumeGroupSnapshotContent` manages `VolumeGroupSnapshot`.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/volumegroupcontroller:tag
```
	
3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/volumegroupcontroller:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Example use case
#### Creating snapshots for volume group

1. Create `VolumeGroup` which has a label selector

```bash
cat << EOF | kubectl apply -f - 
apiVersion: volumegroup.example.com/v1alpha1
kind: VolumeGroup
metadata:
  name: volumegroup1
spec:
  selector:
    matchLabels:
      app: my-app
EOF
```

2. Create PVC with the matching label
```bash
cat << EOF | kubectl apply -f - 
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc1
  labels:
    app: my-app
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
EOF
```

3. Check sc,pv,pvc,vs,vsc (No `VolumeSnapshot` and `VolumeSnapshotContent` exist before the next steps)
```bash
kubectl get sc,pv,pvc,vs,vsc
NAME                                                    PROVISIONER           RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
storageclass.storage.k8s.io/csi-hostpath-sc (default)   hostpath.csi.k8s.io   Delete          Immediate           true                   3m22s

NAME                                                        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM          STORAGECLASS      REASON   AGE
persistentvolume/pvc-750ee5a5-6856-4559-ad64-b842f85bad99   1Gi        RWO            Delete           Bound    default/pvc1   csi-hostpath-sc            7s

NAME                         STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS      AGE
persistentvolumeclaim/pvc1   Bound    pvc-750ee5a5-6856-4559-ad64-b842f85bad99   1Gi        RWO            csi-hostpath-sc   7s
```

4. Create `VolumeGroupSnapshot` for `volumeGroupName` "volumegroup1" 
```bash
cat << EOF | kubectl apply -f - 
apiVersion: volumegroup.example.com/v1alpha1
kind: VolumeGroupSnapshot
metadata:
  name: my-group-snapshot
spec:
  volumeGroupName: volumegroup1
EOF
```

5. Confirm that `VolumeGroupSnapshotContent`, `VolumeSnapshot`, and `VolumeSnapshotContent` for the `VolumeGroupSnapshot` are created
```bash
kubectl get volumegroup,volumegroupsnapshot,volumegroupsnapshotcontent,vs,vsc
NAME                                               AGE
volumegroup.volumegroup.example.com/volumegroup1   14h

NAME                                                            READYTOUSE   VOLUMEGROUP    VOLUMEGROUPSNAPSHOTCONTENT
volumegroupsnapshot.volumegroup.example.com/my-group-snapshot   true         volumegroup1   vgsc-my-group-snapshot

NAME                                                                        READYTOUSE   VOLUMEGROUPSNAPSHOT
volumegroupsnapshotcontent.volumegroup.example.com/vgsc-my-group-snapshot   true         my-group-snapshot

NAME                                                                    READYTOUSE   SOURCEPVC   SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS            SNAPSHOTCONTENT                                    CREATIONTIME   AGE
volumesnapshot.snapshot.storage.k8s.io/vs-vgsc-my-group-snapshot-pvc1   true         pvc1                                1Gi           csi-hostpath-snapclass   snapcontent-c2783e1a-a6bd-4415-b5ed-8c080754f304   5s             6s

NAME                                                                                             READYTOUSE   RESTORESIZE   DELETIONPOLICY   DRIVER                VOLUMESNAPSHOTCLASS      VOLUMESNAPSHOT                   VOLUMESNAPSHOTNAMESPACE   AGE
volumesnapshotcontent.snapshot.storage.k8s.io/snapcontent-c2783e1a-a6bd-4415-b5ed-8c080754f304   true         1073741824    Delete           hostpath.csi.k8s.io   csi-hostpath-snapclass   vs-vgsc-my-group-snapshot-pvc1   default                   5s
```

6. Confirm that `VolumeGroupSnapshot` and `VolumeGroupSnapshotContent` have enough information to manage underlying PVCs and VolumeSnapshots.

```bash
kubectl get volumegroupsnapshot my-group-snapshot -o yaml
apiVersion: volumegroup.example.com/v1alpha1
kind: VolumeGroupSnapshot
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"volumegroup.example.com/v1alpha1","kind":"VolumeGroupSnapshot","metadata":{"annotations":{},"name":"my-group-snapshot","namespace":"default"},"spec":{"volumeGroupName":"volumegroup1"}}
  creationTimestamp: "2022-07-01T17:13:26Z"
  generation: 2
  name: my-group-snapshot
  namespace: default
  resourceVersion: "176322"
  uid: a97bdef3-d0b2-45a4-901a-b16fa4ab8d2a
spec:
  boundVolumeGroupSnapshotContentName: vgsc-my-group-snapshot
  volumeGroupName: volumegroup1
status:
  readyToUse: true


kubectl get volumegroupsnapshotcontent vgsc-my-group-snapshot -o yaml
apiVersion: volumegroup.example.com/v1alpha1
kind: VolumeGroupSnapshotContent
metadata:
  creationTimestamp: "2022-07-01T17:13:26Z"
  generation: 2
  name: vgsc-my-group-snapshot
  namespace: default
  ownerReferences:
  - apiVersion: volumegroup.example.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: VolumeGroupSnapshot
    name: my-group-snapshot
    uid: a97bdef3-d0b2-45a4-901a-b16fa4ab8d2a
  resourceVersion: "176321"
  uid: d47269f5-ef30-4718-ae17-06d977cdef56
spec:
  persistentVolumeClaimList:
  - pvc1
  snapshotList:
  - vs-vgsc-my-group-snapshot-pvc1
  volumeGroupSnapshotName: my-group-snapshot
status:
  readyToUse: true
```

7. Add one more PVC with the matching label

```bash
cat << EOF | kubectl apply -f - 
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc2
  labels:
    app: my-app
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
EOF
```

8. Create `VolumeGroupSnapshot` for the same `VolumeGroup` ("volumegroup1")

```bash
cat << EOF | kubectl apply -f - 
apiVersion: volumegroup.example.com/v1alpha1
kind: VolumeGroupSnapshot
metadata:
  name: my-group-snapshot2
spec:
  volumeGroupName: volumegroup1
EOF
```

9. Confirm that snapshots for the two PVCs are created and managed

```bash
kubectl get volumegroupsnapshot my-group-snapshot2 -o yaml
apiVersion: volumegroup.example.com/v1alpha1
kind: VolumeGroupSnapshot
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"volumegroup.example.com/v1alpha1","kind":"VolumeGroupSnapshot","metadata":{"annotations":{},"name":"my-group-snapshot2","namespace":"default"},"spec":{"volumeGroupName":"volumegroup1"}}
  creationTimestamp: "2022-07-01T17:14:34Z"
  generation: 2
  name: my-group-snapshot2
  namespace: default
  resourceVersion: "176413"
  uid: 81d169fc-47d9-457b-9f45-cfab71edae7f
spec:
  boundVolumeGroupSnapshotContentName: vgsc-my-group-snapshot2
  volumeGroupName: volumegroup1
status:
  readyToUse: true


kubectl get volumegroupsnapshotcontent vgsc-my-group-snapshot2 -o yaml
apiVersion: volumegroup.example.com/v1alpha1
kind: VolumeGroupSnapshotContent
metadata:
  creationTimestamp: "2022-07-01T17:14:34Z"
  generation: 3
  name: vgsc-my-group-snapshot2
  namespace: default
  ownerReferences:
  - apiVersion: volumegroup.example.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: VolumeGroupSnapshot
    name: my-group-snapshot2
    uid: 81d169fc-47d9-457b-9f45-cfab71edae7f
  resourceVersion: "176412"
  uid: 67221ab7-4636-4f36-8ac8-0c00fa96708f
spec:
  persistentVolumeClaimList:
  - pvc1
  - pvc2
  snapshotList:
  - vs-vgsc-my-group-snapshot2-pvc1
  - vs-vgsc-my-group-snapshot2-pvc2
  volumeGroupSnapshotName: my-group-snapshot2
status:
  readyToUse: true

kubectl get vs,vsc
NAME                                                                     READYTOUSE   SOURCEPVC   SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS            SNAPSHOTCONTENT                                    CREATIONTIME   AGE
volumesnapshot.snapshot.storage.k8s.io/vs-vgsc-my-group-snapshot-pvc1    true         pvc1                                1Gi           csi-hostpath-snapclass   snapcontent-c2783e1a-a6bd-4415-b5ed-8c080754f304   2m44s          2m45s
volumesnapshot.snapshot.storage.k8s.io/vs-vgsc-my-group-snapshot2-pvc1   true         pvc1                                1Gi           csi-hostpath-snapclass   snapcontent-d71907de-7b11-48a2-87d1-b932824230dc   96s            97s
volumesnapshot.snapshot.storage.k8s.io/vs-vgsc-my-group-snapshot2-pvc2   true         pvc2                                1Gi           csi-hostpath-snapclass   snapcontent-41f87e5a-5984-47ab-9ab6-3c81ea593605   96s            97s

NAME                                                                                             READYTOUSE   RESTORESIZE   DELETIONPOLICY   DRIVER                VOLUMESNAPSHOTCLASS      VOLUMESNAPSHOT                    VOLUMESNAPSHOTNAMESPACE   AGE
volumesnapshotcontent.snapshot.storage.k8s.io/snapcontent-41f87e5a-5984-47ab-9ab6-3c81ea593605   true         1073741824    Delete           hostpath.csi.k8s.io   csi-hostpath-snapclass   vs-vgsc-my-group-snapshot2-pvc2   default                   96s
volumesnapshotcontent.snapshot.storage.k8s.io/snapcontent-c2783e1a-a6bd-4415-b5ed-8c080754f304   true         1073741824    Delete           hostpath.csi.k8s.io   csi-hostpath-snapclass   vs-vgsc-my-group-snapshot-pvc1    default                   2m44s
volumesnapshotcontent.snapshot.storage.k8s.io/snapcontent-d71907de-7b11-48a2-87d1-b932824230dc   true         1073741824    Delete           hostpath.csi.k8s.io   csi-hostpath-snapclass   vs-vgsc-my-group-snapshot2-pvc1   default                   96s
```

10. Delete the old `VolumeGroupSnapshot` ("my-group-snapshot")

```bash
kubectl delete volumegroupsnapshot my-group-snapshot
```

11. Confirm that all the managed resources are deleted
```bash
kubectl get volumegroup,volumegroupsnapshot,volumegroupsnapshotcontent,vs,vsc
NAME                                               AGE
volumegroup.volumegroup.example.com/volumegroup1   14h

NAME                                                             READYTOUSE   VOLUMEGROUP    VOLUMEGROUPSNAPSHOTCONTENT
volumegroupsnapshot.volumegroup.example.com/my-group-snapshot2   true         volumegroup1   vgsc-my-group-snapshot2

NAME                                                                         READYTOUSE   VOLUMEGROUPSNAPSHOT
volumegroupsnapshotcontent.volumegroup.example.com/vgsc-my-group-snapshot2   true         my-group-snapshot2

NAME                                                                     READYTOUSE   SOURCEPVC   SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS            SNAPSHOTCONTENT                                    CREATIONTIME   AGE
volumesnapshot.snapshot.storage.k8s.io/vs-vgsc-my-group-snapshot2-pvc1   true         pvc1                                1Gi           csi-hostpath-snapclass   snapcontent-d71907de-7b11-48a2-87d1-b932824230dc   20m            20m
volumesnapshot.snapshot.storage.k8s.io/vs-vgsc-my-group-snapshot2-pvc2   true         pvc2                                1Gi           csi-hostpath-snapclass   snapcontent-41f87e5a-5984-47ab-9ab6-3c81ea593605   20m            20m

NAME                                                                                             READYTOUSE   RESTORESIZE   DELETIONPOLICY   DRIVER                VOLUMESNAPSHOTCLASS      VOLUMESNAPSHOT                    VOLUMESNAPSHOTNAMESPACE   AGE
volumesnapshotcontent.snapshot.storage.k8s.io/snapcontent-41f87e5a-5984-47ab-9ab6-3c81ea593605   true         1073741824    Delete           hostpath.csi.k8s.io   csi-hostpath-snapclass   vs-vgsc-my-group-snapshot2-pvc2   default                   20m
volumesnapshotcontent.snapshot.storage.k8s.io/snapcontent-d71907de-7b11-48a2-87d1-b932824230dc   true         1073741824    Delete           hostpath.csi.k8s.io   csi-hostpath-snapclass   vs-vgsc-my-group-snapshot2-pvc1   default                   20m
```

12. Delete `VolumeGroupSnapshot` ("my-group-snapshot2")

```bash
kubectl delete volumegroupsnapshot my-group-snapshot2

kubectl get vgs,vgsc,vs,vsc
No resources found
```

#### Adding preprovisioned snapshot

1. Preprovision `VolumeSnapshot`s

```bash
cat << EOF | kubectl apply -f - 
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: vs1
spec:
  source:
    persistentVolumeClaimName: pvc1
---
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: vs2
spec:
  source:
    persistentVolumeClaimName: pvc2
EOF


kubectl get vgs,vgsc,vs,vsc
NAME                                         READYTOUSE   SOURCEPVC   SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS            SNAPSHOTCONTENT                                    CREATIONTIME   AGE
volumesnapshot.snapshot.storage.k8s.io/vs1   true         pvc1                                1Gi           csi-hostpath-snapclass   snapcontent-b9051fb7-0d8d-4402-8afb-131165154d5d   4s             4s
volumesnapshot.snapshot.storage.k8s.io/vs2   true         pvc2                                1Gi           csi-hostpath-snapclass   snapcontent-ac7768f1-bbbb-42ea-99f4-325999993266   4s             4s

NAME                                                                                             READYTOUSE   RESTORESIZE   DELETIONPOLICY   DRIVER                VOLUMESNAPSHOTCLASS      VOLUMESNAPSHOT   VOLUMESNAPSHOTNAMESPACE   AGE
volumesnapshotcontent.snapshot.storage.k8s.io/snapcontent-ac7768f1-bbbb-42ea-99f4-325999993266   true         1073741824    Delete           hostpath.csi.k8s.io   csi-hostpath-snapclass   vs2              default                   4s
volumesnapshotcontent.snapshot.storage.k8s.io/snapcontent-b9051fb7-0d8d-4402-8afb-131165154d5d   true         1073741824    Delete           hostpath.csi.k8s.io   csi-hostpath-snapclass   vs1              default                   4s
```

2. Create `VolumeGroupSnapshot` with `BoundVolumeGroupSnapshotContentName` and without `volumeGroupName` 
```bash
cat << EOF | kubectl apply -f - 
apiVersion: volumegroup.example.com/v1alpha1
kind: VolumeGroupSnapshot
metadata:
  name: preprovisioned-vgs
spec:
  boundVolumeGroupSnapshotContentName: preprovisioned-vgsc
EOF
```

3. Create `VolumeGroupSnapshotContent` manually with specifying `VolumeGroupSnapshotName` and `SnapshotList`

```bash
cat << EOF | kubectl apply -f - 
apiVersion: volumegroup.example.com/v1alpha1
kind: VolumeGroupSnapshotContent
metadata:
  name: preprovisioned-vgsc
spec:
  volumeGroupSnapshotName: preprovisioned-vgs
  snapshotList: 
    - vs1
    - vs2
EOF
```

4. Confirm that `Status.ReadyToUse` for the `VolumeGroupSnapshot` and the `VolumeGroupSnapshotContent` become true, once all the status of the snapshot managed by them become ready to use
```bash
kubectl get vgs,vgsc,vs,vsc
NAME                                                             READYTOUSE   VOLUMEGROUP   VOLUMEGROUPSNAPSHOTCONTENT
volumegroupsnapshot.volumegroup.example.com/preprovisioned-vgs   true                       preprovisioned-vgsc

NAME                                                                     READYTOUSE   VOLUMEGROUPSNAPSHOT
volumegroupsnapshotcontent.volumegroup.example.com/preprovisioned-vgsc   true         preprovisioned-vgs

NAME                                         READYTOUSE   SOURCEPVC   SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS            SNAPSHOTCONTENT                                    CREATIONTIME   AGE
volumesnapshot.snapshot.storage.k8s.io/vs1   true         pvc1                                1Gi           csi-hostpath-snapclass   snapcontent-b9051fb7-0d8d-4402-8afb-131165154d5d   13m            13m
volumesnapshot.snapshot.storage.k8s.io/vs2   true         pvc2                                1Gi           csi-hostpath-snapclass   snapcontent-ac7768f1-bbbb-42ea-99f4-325999993266   13m            13m

NAME                                                                                             READYTOUSE   RESTORESIZE   DELETIONPOLICY   DRIVER                VOLUMESNAPSHOTCLASS      VOLUMESNAPSHOT   VOLUMESNAPSHOTNAMESPACE   AGE
volumesnapshotcontent.snapshot.storage.k8s.io/snapcontent-ac7768f1-bbbb-42ea-99f4-325999993266   true         1073741824    Delete           hostpath.csi.k8s.io   csi-hostpath-snapclass   vs2              default                   13m
volumesnapshotcontent.snapshot.storage.k8s.io/snapcontent-b9051fb7-0d8d-4402-8afb-131165154d5d   true         1073741824    Delete           hostpath.csi.k8s.io   csi-hostpath-snapclass   vs1              default                   13m
```

5. Delete the `VolumeGroupSnapshot` and the `VolumeGroupSnapshotContent`

```bash
kubectl delete volumeGroupSnapshotContent preprovisioned-vgsc
kubectl delete volumeGroupSnapshot preprovisioned-vgs
```

6. Confirm that snapshots are not deleted

```bash
NAME                                         READYTOUSE   SOURCEPVC   SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS            SNAPSHOTCONTENT                                    CREATIONTIME   AGE
volumesnapshot.snapshot.storage.k8s.io/vs1   true         pvc1                                1Gi           csi-hostpath-snapclass   snapcontent-b9051fb7-0d8d-4402-8afb-131165154d5d   16m            16m
volumesnapshot.snapshot.storage.k8s.io/vs2   true         pvc2                                1Gi           csi-hostpath-snapclass   snapcontent-ac7768f1-bbbb-42ea-99f4-325999993266   16m            16m

NAME                                                                                             READYTOUSE   RESTORESIZE   DELETIONPOLICY   DRIVER                VOLUMESNAPSHOTCLASS      VOLUMESNAPSHOT   VOLUMESNAPSHOTNAMESPACE   AGE
volumesnapshotcontent.snapshot.storage.k8s.io/snapcontent-ac7768f1-bbbb-42ea-99f4-325999993266   true         1073741824    Delete           hostpath.csi.k8s.io   csi-hostpath-snapclass   vs2              default                   16m
volumesnapshotcontent.snapshot.storage.k8s.io/snapcontent-b9051fb7-0d8d-4402-8afb-131165154d5d   true         1073741824    Delete           hostpath.csi.k8s.io   csi-hostpath-snapclass   vs1              default                   16m
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

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

