package dao

import (
	"asyncKubeManager/pkg/dbresolver"
	"asyncKubeManager/pkg/manager/pvc"
	"asyncKubeManager/pkg/model"
	"context"
	"fmt"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
)

func InsertDisk(ctx context.Context, dbResolver *dbresolver.DBResolver, diskName, diskPath string, size int64, vmID int64, pvcID int64) (*model.Disk, error) {
	db := dbResolver.GetDB()

	disk := model.Disk{
		DiskName: diskName,
		Size:     size,
		DiskPath: diskPath,
		VMID:     vmID,
		PVCID:    types.UID(strconv.FormatInt(pvcID, 10)),
	}

	err := db.WithContext(ctx).Create(&disk).Error
	return &disk, err
}

func GetDiskByID(ctx context.Context, dbResolver *dbresolver.DBResolver, id int64) (*model.Disk, error) {
	db := dbResolver.GetDB()
	return GetDiskByIDWithDB(ctx, db, id)
}

func GetDiskByIDWithDB(ctx context.Context, db *gorm.DB, id int64) (*model.Disk, error) {
	d := model.Disk{}
	err := db.WithContext(ctx).Model(&d).Where("id = ?", id).First(&d).Error
	return &d, err
}

func DeleteDiskByID(ctx context.Context, dbResolver *dbresolver.DBResolver, id int64) error {
	db := dbResolver.GetDB()
	return db.WithContext(ctx).Where("id = ?", id).Delete(&model.Disk{}).Error
}

func CreatePVCForVM(ctx context.Context, dbResolver *dbresolver.DBResolver, pvcManager *pvc.K8sPVCManager, vmName string, diskSize string, storageClass string) (*model.Disk, error) {
	db := dbResolver.GetDB()

	pvcName := pvc.GeneratePVCName(vmName)

	exists, err := pvcManager.CheckPVCExists(ctx, "async-km", pvcName)
	if err != nil {
		return nil, fmt.Errorf("error checking PVC existence: %v", err)
	}
	if exists {
		return nil, fmt.Errorf("PVC %s already exists", pvcName)
	}

	pvcResource, err := pvcManager.CreatePVC(ctx, pvcName, diskSize, storageClass)
	if err != nil {
		return nil, fmt.Errorf("error creating PVC: %v", err)
	}

	// 插入磁盘记录到数据库
	quantity := pvcResource.Spec.Resources.Requests[corev1.ResourceStorage]
	disk := model.Disk{
		DiskName: pvcResource.Name,
		Size:     int64(quantity.Value()),
		DiskPath: pvcResource.Spec.VolumeName,
		VMID:     0, // 需要关联VMID
		PVCID:    pvcResource.UID,
	}
	err = db.WithContext(ctx).Create(&disk).Error
	if err != nil {
		return nil, fmt.Errorf("error inserting disk into database: %v", err)
	}

	return &disk, nil
}

func UpdatePVCSize(ctx context.Context, dbResolver *dbresolver.DBResolver, pvcManager *pvc.K8sPVCManager, pvcName, newSize string) (*model.Disk, error) {
	db := dbResolver.GetDB()

	pvcResource, err := pvcManager.ResizePVC(ctx, "async-km", pvcName, newSize)
	if err != nil {
		return nil, fmt.Errorf("error resizing PVC: %v", err)
	}

	disk := model.Disk{}
	err = db.WithContext(ctx).Where("pvc_id = ?", pvcResource.UID).First(&disk).Error
	if err != nil {
		return nil, fmt.Errorf("error finding disk record: %v", err)
	}

	quantity := pvcResource.Spec.Resources.Requests[corev1.ResourceStorage]
	err = db.WithContext(ctx).Model(&disk).Updates(map[string]interface{}{
		"size": int64(quantity.Value()),
	}).Error
	if err != nil {
		return nil, fmt.Errorf("error updating disk record: %v", err)
	}

	return &disk, nil
}

func DeletePVCAndDisk(ctx context.Context, dbResolver *dbresolver.DBResolver, pvcManager *pvc.K8sPVCManager, pvcName string) error {
	db := dbResolver.GetDB()

	err := pvcManager.DeletePVC(ctx, "async-km", pvcName)
	if err != nil {
		return fmt.Errorf("error deleting PVC: %v", err)
	}

	disk := model.Disk{}
	err = db.WithContext(ctx).Where("disk_name = ?", pvcName).First(&disk).Error
	if err != nil {
		return fmt.Errorf("error finding disk record: %v", err)
	}

	err = db.WithContext(ctx).Where("id = ?", disk.ID).Delete(&model.Disk{}).Error
	if err != nil {
		return fmt.Errorf("error deleting disk record: %v", err)
	}

	return nil
}

func ListDisks(ctx context.Context, dbResolver *dbresolver.DBResolver) ([]model.Disk, error) {
	db := dbResolver.GetDB()
	var disks []model.Disk
	err := db.WithContext(ctx).Find(&disks).Error
	return disks, err
}

func UpdateDiskByID(ctx context.Context, dbResolver *dbresolver.DBResolver, pvcManager *pvc.K8sPVCManager, id int64, updates map[string]interface{}) (*model.Disk, error) {
	db := dbResolver.GetDB()

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	disk := model.Disk{}
	err := tx.WithContext(ctx).Where("id = ?", id).First(&disk).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error finding disk: %v", err)
	}

	err = tx.WithContext(ctx).Model(&disk).Updates(updates).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error updating disk: %v", err)
	}

	_, err = pvcManager.ResizePVC(ctx, "async-km", disk.DiskName, updates["size"].(string))
	if err != nil {
		tx.Rollback() // 修改PVC失败，回滚事务
		return nil, fmt.Errorf("error resizing PVC: %v", err)
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &disk, nil
}
