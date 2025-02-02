package pvc

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PVCManager defines the interface for managing PVC resources.
type PVCManager interface {
	CreatePVC(ctx context.Context, pvcName string, diskSize string, storageClass string) (*corev1.PersistentVolumeClaim, error)
	DeletePVC(ctx context.Context, namespace, name string) error
	UpdatePVC(ctx context.Context, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error)
	ResizePVC(ctx context.Context, namespace, name, newSize string) (*corev1.PersistentVolumeClaim, error)
}

// K8sPVCManager implements the PVCManager interface using the Kubernetes client.
type K8sPVCManager struct {
	Client kubernetes.Interface
}

// CreatePVC creates a new PVC resource in Kubernetes.
func (m *K8sPVCManager) CreatePVC(ctx context.Context, pvcName string, diskSize string, storageClass string) (*corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: "async-km",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(diskSize),
				},
			},
			StorageClassName: &storageClass,
		},
	}
	return m.Client.CoreV1().PersistentVolumeClaims("async-km").Create(ctx, pvc, metav1.CreateOptions{})
}

// GetPVCByName retrieves a PVC resource by its name.
func (m *K8sPVCManager) GetPVCByName(ctx context.Context, namespace, name string) (*corev1.PersistentVolumeClaim, error) {
	return m.Client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetPVCByUID retrieves a PVC resource by its UID.
func (m *K8sPVCManager) GetPVCByUID(ctx context.Context, namespace, uid string) (*corev1.PersistentVolumeClaim, error) {
	pvcs, err := m.Client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.uid=%s", uid),
	})
	if err != nil {
		return nil, err
	}

	if len(pvcs.Items) == 0 {
		return nil, fmt.Errorf("PVC with UID %s not found", uid)
	}

	return &pvcs.Items[0], nil
}

// UpdatePVC updates an existing PVC resource in Kubernetes.
func (m *K8sPVCManager) UpdatePVC(ctx context.Context, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	return m.Client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Update(ctx, pvc, metav1.UpdateOptions{})
}

// DeletePVC deletes a PVC resource in Kubernetes.
func (m *K8sPVCManager) DeletePVC(ctx context.Context, namespace, name string) error {
	return m.Client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// ResizePVC resizes an existing PVC resource in Kubernetes.
func (m *K8sPVCManager) ResizePVC(ctx context.Context, namespace, name, newSize string) (*corev1.PersistentVolumeClaim, error) {
	pvc, err := m.GetPVCByName(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	pvc.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse(newSize)
	return m.Client.CoreV1().PersistentVolumeClaims(namespace).Update(ctx, pvc, metav1.UpdateOptions{})
}

// CheckPVCExists checks if a PVC exists in the specified namespace.
func (m *K8sPVCManager) CheckPVCExists(ctx context.Context, namespace, name string) (bool, error) {
	_, err := m.GetPVCByName(ctx, namespace, name)
	if err != nil {
		return false, err
	}
	return true, nil
}
