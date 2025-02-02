package pvc

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GeneratePVCName generates a PVC name based on the VM name.
func GeneratePVCName(vmName string) string {
	return fmt.Sprintf("%s-disk", vmName)
}

// CheckPVCExists checks if a PVC exists in the specified namespace.
func CheckPVCExists(ctx context.Context, client kubernetes.Interface, namespace, name string) (bool, error) {
	_, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}
