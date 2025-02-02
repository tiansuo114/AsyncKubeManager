package model

import "k8s.io/apimachinery/pkg/types"

type VM struct {
	ID        int64   `gorm:"primary_key;AUTO_INCREMENT"`
	VMName    string  `gorm:"not null; index:vm_name; type:varchar(32)"`
	CPU       int64   `gorm:"not null; index:cpu;"`
	Memory    int64   `gorm:"not null; index:memory;"`
	Storage   int64   `gorm:"not null; index:storage;"`
	DiskIDs   []int64 `gorm:"not null; index:disk_ids;"`
	Disks     []Disk  `gorm:"-"`
	CreatedAt int64   `gorm:"autoCreateTime:milli; not null"`
	UpdatedAt int64   `gorm:"autoUpdateTime:milli; not null"`
}

type Disk struct {
	ID        int64     `gorm:"primary_key;AUTO_INCREMENT"`
	DiskName  string    `gorm:"not null; index:disk_name; type:varchar(32)"`
	Size      int64     `gorm:"not null; index:size;"`
	DiskPath  string    `gorm:"not null; index:disk_path; type:varchar(32)"`
	VMID      int64     `gorm:"not null; index:vm_id"`
	PVCID     types.UID `gorm:"not null; index:pvc_id"`
	CreatedAt int64     `gorm:"autoCreateTime:milli; not null"`
	UpdatedAt int64     `gorm:"autoUpdateTime:milli; not null"`
}
