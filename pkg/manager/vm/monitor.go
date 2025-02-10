package vm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// VMMonitor defines the interface for monitoring VirtualMachine resources.
type VMMonitor interface {
	MonitorVMCreation(ctx context.Context, namespace, name string, timeout time.Duration) error
	MonitorVMs(ctx context.Context, namespace string, vmNames *[]string, interval time.Duration, wg *sync.WaitGroup)
}

// KubevirtVMMonitor implements the VMMonitor interface using the Kubernetes kubeVirtClientSet.
type KubevirtVMMonitor struct {
	Client *kubernetes.Clientset
}

// MonitorVMCreation monitors the VirtualMachine creation status and triggers an alert if not created within the timeout.
func (m *KubevirtVMMonitor) MonitorVMCreation(ctx context.Context, namespace, name string, timeout time.Duration) error {
	startTime := time.Now()
	for {
		vm, err := m.Client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if vm.Status.Phase == corev1.PodRunning {
			return nil
		}

		if time.Since(startTime) > timeout {
			return fmt.Errorf("VirtualMachine %s/%s not created within %v", namespace, name, timeout)
		}

		time.Sleep(5 * time.Second)
	}
}

// MonitorVMs monitors a list of VirtualMachines and checks their status periodically.
func (m *KubevirtVMMonitor) MonitorVMs(ctx context.Context, namespace string, vmNames *[]string, interval time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, name := range *vmNames {
				vm, err := m.Client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
				if err != nil {
					zap.L().Error(fmt.Sprintf("Error getting VirtualMachine %s/%s: %v\n", namespace, name, err))
					continue
				}

				switch vm.Status.Phase {
				case corev1.PodRunning:
					zap.L().Info(fmt.Sprintf("VirtualMachine %s/%s is running\n", namespace, name))
				case corev1.PodPending:
					zap.L().Info(fmt.Sprintf("VirtualMachine %s/%s is pending\n", namespace, name))
				case corev1.PodFailed:
					zap.L().Info(fmt.Sprintf("VirtualMachine %s/%s has failed\n", namespace, name))
				default:
					zap.L().Info(fmt.Sprintf("VirtualMachine %s/%s has an unknown status: %s\n", namespace, name, vm.Status.Phase))
				}
			}
		case <-ctx.Done():
			zap.L().Info(fmt.Sprintf("Stopping VirtualMachine monitoring"))
			return
		}
	}
}
