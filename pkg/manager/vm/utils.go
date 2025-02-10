package vm

import (
	"asyncKubeManager/pkg/model"
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GenerateVMNameFromVMModel generates a VirtualMachine name based on the given prefix.
func GenerateVMNameFromVMModel(vm *model.VM) string {
	return fmt.Sprintf("%s-%s", vm.VMName, vm.UID)
}

// CheckVMExists checks if a VirtualMachine exists in the specified namespace.
func CheckVMExists(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (bool, error) {
	_, err := client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}

// GenerateDataValumName generates a PVC name based on the VM name.
func GenerateDataValumName(vmName string) string {
	return fmt.Sprintf("%s-dv", vmName)
}
