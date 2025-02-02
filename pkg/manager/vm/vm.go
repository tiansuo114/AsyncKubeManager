package vm

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubevirt"
)

// VMManager defines the interface for managing VirtualMachine resources.
type VMManager interface {
	CreateVM(ctx context.Context, namespace string, vm *kubevirtv1.VirtualMachine) (*kubevirtv1.VirtualMachine, error)
	DeleteVM(ctx context.Context, namespace, name string) error
	GetVM(ctx context.Context, namespace, name string) (*kubevirtv1.VirtualMachine, error)
	UpdateVM(ctx context.Context, namespace string, vm *kubevirtv1.VirtualMachine) (*kubevirtv1.VirtualMachine, error)
	ListVMs(ctx context.Context, namespace string) (*kubevirtv1.VirtualMachineList, error)
}

// KubevirtVMManager implements the VMManager interface using the Kubevirt client.
type KubevirtVMManager struct {
	client *kubevirt.Clientset
}

// NewKubevirtVMManager creates a new KubevirtVMManager.
func NewKubevirtVMManager(client *kubevirt.Clientset) *KubevirtVMManager {
	return &KubevirtVMManager{
		client: client,
	}
}

// CreateVM creates a new VirtualMachine resource.
func (m *KubevirtVMManager) CreateVM(ctx context.Context, namespace string, vm *kubevirtv1.VirtualMachine) (*kubevirtv1.VirtualMachine, error) {
	return m.client.KubevirtV1().VirtualMachines(namespace).Create(ctx, vm, metav1.CreateOptions{})
}

// DeleteVM deletes a VirtualMachine resource.
func (m *KubevirtVMManager) DeleteVM(ctx context.Context, namespace, name string) error {
	return m.client.KubevirtV1().VirtualMachines(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetVM retrieves a VirtualMachine resource.
func (m *KubevirtVMManager) GetVM(ctx context.Context, namespace, name string) (*kubevirtv1.VirtualMachine, error) {
	return m.client.KubevirtV1().VirtualMachines(namespace).Get(ctx, name, metav1.GetOptions{})
}

// UpdateVM updates an existing VirtualMachine resource.
func (m *KubevirtVMManager) UpdateVM(ctx context.Context, namespace string, vm *kubevirtv1.VirtualMachine) (*kubevirtv1.VirtualMachine, error) {
	return m.client.KubevirtV1().VirtualMachines(namespace).Update(ctx, vm, metav1.UpdateOptions{})
}

// ListVMs lists all VirtualMachine resources in the specified namespace.
func (m *KubevirtVMManager) ListVMs(ctx context.Context, namespace string) (*kubevirtv1.VirtualMachineList, error) {
	return m.client.KubevirtV1().VirtualMachines(namespace).List(ctx, metav1.ListOptions{})
}
