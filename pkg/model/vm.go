package model

import "gorm.io/gorm"

type VM struct {
	ID        int64    `gorm:"primary_key;AUTO_INCREMENT"` // Primary key
	UID       string   `gorm:"not null; index:hash_id;"`
	VMName    string   `gorm:"not null; index:vm_name; type:varchar(32)"` // Virtual machine name
	CPU       int64    `gorm:"not null; index:cpu;"`                      // CPU cores
	Memory    int64    `gorm:"not null; index:memory;"`                   // Memory size (in MB)
	DiskIDs   []int64  `gorm:"not null; index:disk_ids;"`                 // List of associated disk IDs
	Disks     []Disk   `gorm:"-"`                                         // Associated disks (not stored in DB)
	DVID      string   `gorm:"not null;"`
	DVName    string   `gorm:"not null;"`
	OsMirror  string   `gorm:"not null;"`
	Os        OSMirror `gorm:"-"`
	Status    VMStatus `gorm:"not null; type:varchar(32); index:status;"`            // VM status
	CreatedAt int64    `gorm:"autoCreateTime:milli; not null; index:idx_created_at"` // Creation time
	Creator   string   `gorm:"not null; type:varchar(32)"`                           // Creator
	UpdatedAt int64    `gorm:"autoUpdateTime:milli; not null"`                       // Update time
	Updater   string   `gorm:"not null; type:varchar(32)"`                           // Updater

	gorm.DeletedAt // Soft delete field
}

type VMStatus string

const (
	VMStatusPendingCreation VMStatus = "PendingCreation"
	VMStatusPendingStart    VMStatus = "PendingStart"
	VMStatusRunning         VMStatus = "Running"
	VMStatusPendingStop     VMStatus = "PendingStop"
	VMStatusStopped         VMStatus = "Stopped"
	VMStatusPendingDeletion VMStatus = "PendingDeletion"
	VMStatusDeleted         VMStatus = "Deleted"
	VMStatusError           VMStatus = "Error"
)

func (VM) TableName() string {
	return "vm"
}
