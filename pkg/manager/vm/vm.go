package vm

import (
	"asyncKubeManager/cmd/console/app/options"
	"asyncKubeManager/pkg/dbresolver"
	"asyncKubeManager/pkg/manager/pvc"
	"asyncKubeManager/pkg/token"
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kubevirtv1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubevirt"

	cdiCli "kubevirt.io/client-go/containerizeddataimporter"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

// VmManager defines the interface for managing VirtualMachine resources.
type VmManager interface {
	CreateVM(ctx context.Context, vm *kubevirtv1.VirtualMachine) (*kubevirtv1.VirtualMachine, error)
	DeleteVM(ctx context.Context, name string) error
	GetVM(ctx context.Context, name string) (*kubevirtv1.VirtualMachine, error)
	UpdateVM(ctx context.Context, vm *kubevirtv1.VirtualMachine) (*kubevirtv1.VirtualMachine, error)
	ListVMs(ctx context.Context) (*kubevirtv1.VirtualMachineList, error)

	// CheckVMExists checks if a VirtualMachine exists in the specified .
	CheckVMExists(ctx context.Context, name string) (bool, error)
	// GetVMByUID retrieves a VirtualMachine resource by its UID.
	GetVMByUID(ctx context.Context, uid types.UID) (*kubevirtv1.VirtualMachine, error)
	// StartVM sets spec.running to true to start the VM.
	StartVM(ctx context.Context, name string) (*kubevirtv1.VirtualMachine, error)
	// StopVM sets spec.running to false to stop the VM.
	StopVM(ctx context.Context, name string) (*kubevirtv1.VirtualMachine, error)
	// RestartVM triggers a restart by patching an annotation.
	RestartVM(ctx context.Context, name string) (*kubevirtv1.VirtualMachine, error)
	// PatchVM applies a generic patch to the VirtualMachine.
	PatchVM(ctx context.Context, name string, patchData []byte) (*kubevirtv1.VirtualMachine, error)
}

// KubevirtVMManager implements the VmManager interface using the KubeVirt kubeVirtClientSet.
type KubevirtVMManager struct {
	kubeVirtClientSet *kubevirt.Clientset
	cdiClientSet      *cdiCli.Clientset
	dbResolver        *dbresolver.DBResolver
	pvcManager        *pvc.K8sPVCManager
}

// NewKubevirtVMManager creates a new KubevirtVMManager.
func NewKubevirtVMManager(kubeVirtClientSet *kubevirt.Clientset, cdiClientSet *cdiCli.Clientset, dbResolver *dbresolver.DBResolver, pvcManager *pvc.K8sPVCManager) *KubevirtVMManager {
	return &KubevirtVMManager{
		kubeVirtClientSet: kubeVirtClientSet,
		cdiClientSet:      cdiClientSet,
		dbResolver:        dbResolver,
		pvcManager:        pvcManager,
	}
}

// 将上述参数填入函数的传参列表中
func (m *KubevirtVMManager) CreateVM(ctx context.Context, vmname string, cpu int64,
	memory int64, storage int64, osMirrorUrl string) (*cdiv1.DataVolume, *kubevirtv1.VirtualMachine, error) {
	dv, err := m.CreateDataVolumeForVM(ctx, vmname, fmt.Sprintf("%dGi", storage), osMirrorUrl)
	if err != nil {
		return nil, nil, err
	}

	vm := &kubevirtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vmname,
			Namespace: options.S.K8sNameSpace,
			Annotations: map[string]string{
				"creator": token.GetNameFromCtx(ctx),
				"updater": token.GetNameFromCtx(ctx),
			},
		},
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"vmName": vmname,
					},
				},
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{
						CPU: &kubevirtv1.CPU{
							// 将数据库中记录的 CPU 核数赋值
							Cores: uint32(cpu),
						},
						Memory: &kubevirtv1.Memory{
							Guest:    resource.NewQuantity(memory, resource.BinarySI),
							MaxGuest: resource.NewQuantity(memory*2, resource.BinarySI),
						},
						Devices: kubevirtv1.Devices{
							Disks: []kubevirtv1.Disk{
								{
									Name: dv.Name,
									DiskDevice: kubevirtv1.DiskDevice{
										Disk: &kubevirtv1.DiskTarget{
											Bus:      "virtio",
											ReadOnly: false,
										},
									},
								},
							},
						},
					},
					Volumes: []kubevirtv1.Volume{
						{
							Name: dv.Name,
							VolumeSource: kubevirtv1.VolumeSource{
								DataVolume: &kubevirtv1.DataVolumeSource{
									Name:         dv.Name,
									Hotpluggable: false,
								},
							},
						},
					},
				},
			},
		},
	}
	// 通过 KubeVirt 客户端创建 VirtualMachine 资源
	resVM, err := m.kubeVirtClientSet.KubevirtV1().VirtualMachines(options.S.K8sNameSpace).Create(ctx, vm, metav1.CreateOptions{})
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		// 如果在函数执行过程中有任何错误，进入该分支
		if err != nil {
			// 如果 VirtualMachine 已创建，则尝试删除它
			if resVM != nil {
				if delErr := m.DeleteVM(ctx, resVM.Name); delErr != nil {
					zap.L().Error(fmt.Sprintf("删除 VirtualMachine %s 失败: %v", vm.Name, delErr))
				} else {
					zap.L().Error(fmt.Sprintf("已删除 VirtualMachine %s", vm.Name))
				}
			}

			// 如果 DataVolume 已创建，则尝试删除它
			if dv != nil {
				if delErr := m.DeleteDataVolume(ctx, dv.Name); delErr != nil {
					zap.L().Error(fmt.Sprintf("删除 DataVolume %s 失败: %v", dv.Name, delErr))
				} else {
					zap.L().Error(fmt.Sprintf("已删除 DataVolume %s", dv.Name))
				}
			}
		}
	}()

	return dv, resVM, err
}

// DeleteVM deletes a VirtualMachine resource.
func (m *KubevirtVMManager) DeleteVM(ctx context.Context, name string) error {
	return m.kubeVirtClientSet.KubevirtV1().VirtualMachines(options.S.K8sNameSpace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetVM retrieves a VirtualMachine resource.
func (m *KubevirtVMManager) GetVM(ctx context.Context, name string) (*kubevirtv1.VirtualMachine, error) {
	return m.kubeVirtClientSet.KubevirtV1().VirtualMachines(options.S.K8sNameSpace).Get(ctx, name, metav1.GetOptions{})
}

// UpdateVM updates an existing VirtualMachine resource.
func (m *KubevirtVMManager) UpdateVM(ctx context.Context, vm *kubevirtv1.VirtualMachine) (*kubevirtv1.VirtualMachine, error) {
	return m.kubeVirtClientSet.KubevirtV1().VirtualMachines(options.S.K8sNameSpace).Update(ctx, vm, metav1.UpdateOptions{})
}

// ListVMs lists all VirtualMachine resources in the specified .
func (m *KubevirtVMManager) ListVMs(ctx context.Context) (*kubevirtv1.VirtualMachineList, error) {
	return m.kubeVirtClientSet.KubevirtV1().VirtualMachines(options.S.K8sNameSpace).List(ctx, metav1.ListOptions{})
}

// CheckVMExists checks if a VirtualMachine exists in the specified .
func (m *KubevirtVMManager) CheckVMExists(ctx context.Context, name string) (bool, error) {
	_, err := m.GetVM(ctx, name)
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetVMByUID retrieves a VirtualMachine resource by its UID.
func (m *KubevirtVMManager) GetVMByUID(ctx context.Context, uid types.UID) (*kubevirtv1.VirtualMachine, error) {
	vms, err := m.kubeVirtClientSet.KubevirtV1().VirtualMachines(options.S.K8sNameSpace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.uid=%s", uid),
	})
	if err != nil {
		return nil, err
	}

	if len(vms.Items) == 0 {
		return nil, fmt.Errorf("VirtualMachine with UID %s not found", uid)
	}

	return &vms.Items[0], nil
}

// StartVM sets spec.running to true to start the VirtualMachine.
func (m *KubevirtVMManager) StartVM(ctx context.Context, name string) (*kubevirtv1.VirtualMachine, error) {
	vm, err := m.GetVM(ctx, name)
	if err != nil {
		return nil, err
	}

	// 如果已经停止，则直接返回
	if vm.Spec.RunStrategy != nil && *vm.Spec.RunStrategy == kubevirtv1.RunStrategyAlways {
		return vm, nil
	}

	*vm.Spec.RunStrategy = kubevirtv1.RunStrategyAlways
	return m.UpdateVM(ctx, vm)
}

// StopVM sets spec.running to false to stop the VirtualMachine.
func (m *KubevirtVMManager) StopVM(ctx context.Context, name string) (*kubevirtv1.VirtualMachine, error) {
	vm, err := m.GetVM(ctx, name)
	if err != nil {
		return nil, err
	}

	// 如果已经停止，则直接返回
	if vm.Spec.RunStrategy != nil && *vm.Spec.RunStrategy == kubevirtv1.RunStrategyHalted {
		return vm, nil
	}

	*vm.Spec.RunStrategy = kubevirtv1.RunStrategyHalted
	return m.UpdateVM(ctx, vm)
}

// RestartVM triggers a restart by patching an annotation with the current timestamp.
func (m *KubevirtVMManager) RestartVM(ctx context.Context, name string) (*kubevirtv1.VirtualMachine, error) {
	patchData := []byte(fmt.Sprintf(`{"metadata": {"annotations": {"kubevirt.io/restart": "%s"}}}`, time.Now().Format(time.RFC3339)))
	return m.kubeVirtClientSet.KubevirtV1().VirtualMachines(options.S.K8sNameSpace).Patch(ctx, name, types.MergePatchType, patchData, metav1.PatchOptions{})
}

// PatchVM applies a generic patch to the VirtualMachine.
func (m *KubevirtVMManager) PatchVM(ctx context.Context, name string, patchData []byte) (*kubevirtv1.VirtualMachine, error) {
	return m.kubeVirtClientSet.KubevirtV1().VirtualMachines(options.S.K8sNameSpace).Patch(ctx, name, types.MergePatchType, patchData, metav1.PatchOptions{})
}

func (m *KubevirtVMManager) CreateDataVolumeForVM(ctx context.Context, vmName string, diskSize string, osMirrorUrl string) (*cdiv1.DataVolume, error) {
	// Generate PVC name based on VM name
	pvcName := GenerateDataValumName(vmName)

	// Check if PVC already exists
	exists, err := m.pvcManager.CheckPVCExists(ctx, options.S.K8sNameSpace, pvcName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("PVC %s already exists", pvcName)
	}

	v := cdiv1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: options.S.K8sNameSpace,
		},
		Spec: cdiv1.DataVolumeSpec{
			Source: &cdiv1.DataVolumeSource{
				Registry: &cdiv1.DataVolumeSourceRegistry{
					URL: &osMirrorUrl,
				},
			},
			PVC: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(diskSize),
					},
				},
			},
		},
	}

	return m.cdiClientSet.CdiV1beta1().DataVolumes(options.S.K8sNameSpace).Create(ctx, &v, metav1.CreateOptions{})
}

func (m *KubevirtVMManager) DeleteDataVolume(ctx context.Context, name string) error {
	return m.cdiClientSet.CdiV1beta1().DataVolumes(options.S.K8sNameSpace).Delete(ctx, name, metav1.DeleteOptions{})
}
