package dao

import (
	"context"
	"gorm.io/gorm"
	"time"

	"asyncKubeManager/pkg/dbresolver"
	"asyncKubeManager/pkg/model"
	"asyncKubeManager/pkg/token"
	"asyncKubeManager/pkg/utils"
)

func InsertUser(ctx context.Context, dbResolver *dbresolver.DBResolver, username, tel, email, desc string) (*model.User, error) {
	db := dbResolver.GetDB()

	creator := token.GetUIDFromCtx(ctx)
	user := model.User{
		UID:      utils.NextID(),
		Username: username,
		Role:     model.Normal,
		Primary:  false,
		Tel:      tel,
		Email:    email,
		Desc:     desc,
		Status:   model.UserStatusEnabled,
		Creator:  creator,
		Updater:  creator,
	}

	err := db.WithContext(ctx).Create(&user).Error
	return &user, err
}

func GetUserByUID(ctx context.Context, dbResolver *dbresolver.DBResolver, uid string) (*model.User, error) {
	db := dbResolver.GetDB()
	return GetUserByUIDWithDB(ctx, db, uid)
}

func GetUserByUIDWithDB(ctx context.Context, db *gorm.DB, uid string) (*model.User, error) {
	u := model.User{}
	err := db.WithContext(ctx).Model(&u).Where("uid = ?", uid).First(&u).Error
	return &u, err
}

func GetUserByUserName(ctx context.Context, dbResolver *dbresolver.DBResolver, username string) (*model.User, error) {
	db := dbResolver.GetDB()
	return GetUserByUserNameWithDB(ctx, db, username)
}

func GetUserByUserNameWithDB(ctx context.Context, db *gorm.DB, username string) (*model.User, error) {
	u := model.User{}
	err := db.WithContext(ctx).Model(&u).Where("username = ?", username).First(&u).Error
	return &u, err
}

func DeleteUserByID(ctx context.Context, dbResolver *dbresolver.DBResolver, uid string) error {
	db := dbResolver.GetDB()
	return db.WithContext(ctx).Where("uid = ?", uid).Delete(&model.User{}).Error
}

func UpdateUserByID(ctx context.Context, dbResolver *dbresolver.DBResolver, uid string, updates map[string]interface{}) error {
	db := dbResolver.GetDB()
	return UpdateUserByUIDWithDB(ctx, db, uid, updates)
}

func UpdateUserByUIDWithDB(ctx context.Context, db *gorm.DB, uid string, updates map[string]interface{}) error {
	updates["updater"] = token.GetUIDFromCtx(ctx)
	updates["updated_at"] = time.Now().UnixMilli()

	return db.WithContext(ctx).Model(&model.User{}).Where("uid = ?", uid).Updates(updates).Error
}

func FindUserOperatorLogsByUid(ctx context.Context, dbResolver *dbresolver.DBResolver, uid string) ([]model.UserOperatorLog, error) {
	db := dbResolver.GetDB()
	var logs []model.UserOperatorLog
	err := db.WithContext(ctx).Where("uid = ?", uid).Find(&logs).Error
	return logs, err
}

func ListUsers(ctx context.Context, dbResolver *dbresolver.DBResolver) ([]model.User, error) {
	db := dbResolver.GetDB()
	var users []model.User
	err := db.WithContext(ctx).Find(&users).Error
	return users, err
}

func ChangeUserRole(ctx context.Context, dbResolver *dbresolver.DBResolver, uid string, role model.UserRole) error {
	return UpdateUserByID(ctx, dbResolver, uid, map[string]interface{}{
		"role": role,
	})
}
