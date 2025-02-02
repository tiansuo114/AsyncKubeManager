package pvc

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sync"
	"time"
)

// PVCMonitor defines the interface for monitoring PVC resources.
type PVCMonitor interface {
	MonitorPVCBinding(ctx context.Context, namespace, name string, timeout time.Duration) error
}

// K8sPVCMonitor implements the PVCMonitor interface using the Kubernetes client.
type K8sPVCMonitor struct {
	Client kubernetes.Interface
}

// MonitorPVCBinding monitors the PVC binding status and triggers an alert if not bound within the timeout.
func (m *K8sPVCMonitor) MonitorPVCBinding(ctx context.Context, namespace, name string, timeout time.Duration) error {
	startTime := time.Now()
	for {
		pvc, err := m.Client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if pvc.Status.Phase == corev1.ClaimBound {
			return nil
		}

		if time.Since(startTime) > timeout {
			return fmt.Errorf("PVC %s/%s not bound within %v", namespace, name, timeout)
		}

		time.Sleep(5 * time.Second)
	}
}

// MonitorPVCs monitors a list of PVCs and checks their status periodically.
func (m *K8sPVCMonitor) MonitorPVCs(ctx context.Context, namespace string, pvcNames *[]string, interval time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, name := range *pvcNames {
				pvc, err := m.Client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
				if err != nil {
					zap.L().Error(fmt.Sprintf("Error getting PVC %s/%s: %v\n", namespace, name, err))
					continue
				}

				switch pvc.Status.Phase {
				case corev1.ClaimBound:
					zap.L().Info(fmt.Sprintf("PVC %s/%s is bound\n", namespace, name))
				case corev1.ClaimPending:
					zap.L().Info(fmt.Sprintf("PVC %s/%s is pending\n", namespace, name))
				case corev1.ClaimLost:
					zap.L().Info(fmt.Sprintf("PVC %s/%s is lost\n", namespace, name))
				default:
					zap.L().Info(fmt.Sprintf("PVC %s/%s has an unknown status: %s\n", namespace, name, pvc.Status.Phase))
				}
			}
		case <-ctx.Done():
			zap.L().Info(fmt.Sprintf("Stopping PVC monitoring"))
			return
		}
	}
}
