package dao

import (
	"context"
	"errors"
	"time"

	"asyncKubeManager/pkg/dbresolver"
	"asyncKubeManager/pkg/model"
	"asyncKubeManager/pkg/token"
	"gorm.io/gorm"
)

// InsertVM inserts a new VM record into the database.
func InsertVM(ctx context.Context, dbResolver *dbresolver.DBResolver, vmName, osMirror, uid string, cpu int64, memory int64) (*model.VM, error) {
	db := dbResolver.GetDB()

	creator := token.GetUIDFromCtx(ctx)
	vm := model.VM{
		UID:       uid,
		VMName:    vmName,
		CPU:       cpu,
		Memory:    memory,
		CreatedAt: time.Now().UnixMilli(),
		Creator:   creator,
		UpdatedAt: time.Now().UnixMilli(),
		Updater:   creator,
		OsMirror:  osMirror,
	}

	err := db.WithContext(ctx).Create(&vm).Error
	return &vm, err
}

// GetVMByID retrieves a VM record by its ID.
func GetVMByID(ctx context.Context, dbResolver *dbresolver.DBResolver, id int64) (bool, *model.VM, error) {
	db := dbResolver.GetDB()
	return GetVMByIDWithDB(ctx, db, id)
}

func GetVMByIDWithDB(ctx context.Context, db *gorm.DB, id int64) (bool, *model.VM, error) {
	vm := model.VM{}
	err := db.WithContext(ctx).Where("id = ?", id).First(&vm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &vm, nil
}

// GetVMByName retrieves a VM record by its name.
func GetVMByName(ctx context.Context, dbResolver *dbresolver.DBResolver, vmName string) (bool, *model.VM, error) {
	db := dbResolver.GetDB()
	return GetVMByNameWithDB(ctx, db, vmName)
}

func GetVMByNameWithDB(ctx context.Context, db *gorm.DB, vmName string) (bool, *model.VM, error) {
	vm := model.VM{}
	err := db.WithContext(ctx).Where("vm_name = ?", vmName).First(&vm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &vm, nil
}

// UpdateVMByID updates the VM record with the specified ID.
func UpdateVMByID(ctx context.Context, dbResolver *dbresolver.DBResolver, id int64, updates map[string]interface{}) error {
	db := dbResolver.GetDB()
	updates["updater"] = token.GetUIDFromCtx(ctx)
	updates["updated_at"] = time.Now().UnixMilli()

	return db.WithContext(ctx).Model(&model.VM{}).Where("id = ?", id).Updates(updates).Error
}

func UpdateVMByUID(ctx context.Context, dbResolver *dbresolver.DBResolver, uid string, updates map[string]interface{}) error {
	db := dbResolver.GetDB()
	updates["updater"] = token.GetUIDFromCtx(ctx)
	updates["updated_at"] = time.Now().UnixMilli()

	return db.WithContext(ctx).Model(&model.VM{}).Where("uid = ?", uid).Updates(updates).Error
}

func UpdateVMByName(ctx context.Context, dbResolver *dbresolver.DBResolver, vmName string, updates map[string]interface{}) error {
	db := dbResolver.GetDB()
	updates["updater"] = token.GetUIDFromCtx(ctx)
	updates["updated_at"] = time.Now().UnixMilli()

	return db.WithContext(ctx).Model(&model.VM{}).Where("vm_name = ?", vmName).Updates(updates).Error
}

// DeleteVMByID deletes a VM record by its ID.
func DeleteVMByID(ctx context.Context, dbResolver *dbresolver.DBResolver, id int64) error {
	db := dbResolver.GetDB()
	return db.WithContext(ctx).Where("id = ?", id).Delete(&model.VM{}).Error
}

// AddDiskToVM associates a disk with a VM.
func AddDiskToVM(ctx context.Context, dbResolver *dbresolver.DBResolver, vmID int64, diskID int64) error {
	db := dbResolver.GetDB()
	return db.WithContext(ctx).Model(&model.VM{}).Where("id = ?", vmID).Update("disk_ids", gorm.Expr("array_append(disk_ids, ?)", diskID)).Error
}

// RemoveDiskFromVM removes a disk association from a VM.
func RemoveDiskFromVM(ctx context.Context, dbResolver *dbresolver.DBResolver, vmID int64, diskID int64) error {
	db := dbResolver.GetDB()
	return db.WithContext(ctx).Model(&model.VM{}).Where("id = ?", vmID).Update("disk_ids", gorm.Expr("array_remove(disk_ids, ?)", diskID)).Error
}

// ListVMs retrieves all VM records from the database.
func ListVMs(ctx context.Context, dbResolver *dbresolver.DBResolver) ([]model.VM, error) {
	db := dbResolver.GetDB()
	var vms []model.VM
	err := db.WithContext(ctx).Find(&vms).Error
	return vms, err
}

func ListVMsByOwnerID(ctx context.Context, dbResolver *dbresolver.DBResolver) ([]model.VM, error) {
	db := dbResolver.GetDB()
	var vms []model.VM
	err := db.WithContext(ctx).Where("creator = ?", token.GetUIDFromCtx(ctx)).Find(&vms).Error
	return vms, err
}
