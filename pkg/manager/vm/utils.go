package vm

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GenerateVMName generates a VirtualMachine name based on the given prefix.
func GenerateVMName(prefix string) string {
	return fmt.Sprintf("%s-vm", prefix)
}

// CheckVMExists checks if a VirtualMachine exists in the specified namespace.
func CheckVMExists(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (bool, error) {
	_, err := client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}
