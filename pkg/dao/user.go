package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"

	"asyncKubeManager/pkg/dbresolver"
	"asyncKubeManager/pkg/model"
	"asyncKubeManager/pkg/token"
)

func InsertUser(ctx context.Context, dbResolver *dbresolver.DBResolver, uid, username, tel, email, desc string, role model.UserRole) (*model.User, error) {
	db := dbResolver.GetDB()
	return InsertUserWithDB(ctx, db, uid, username, tel, email, desc, role)
}

func InsertUserWithDB(ctx context.Context, db *gorm.DB, uid, username, tel, email, desc string, role model.UserRole) (*model.User, error) {
	creator := token.GetUIDFromCtx(ctx)
	user := model.User{
		UID:      uid,
		Username: username,
		Role:     role,
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

func GetUserByUID(ctx context.Context, dbResolver *dbresolver.DBResolver, uid string) (bool, *model.User, error) {
	db := dbResolver.GetDB()
	return GetUserByUIDWithDB(ctx, db, uid)
}

func GetUserByUIDWithDB(ctx context.Context, db *gorm.DB, uid string) (bool, *model.User, error) {
	u := model.User{}
	err := db.WithContext(ctx).Model(&u).Where("uid = ?", uid).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &u, err
}

func GetUserByUserName(ctx context.Context, dbResolver *dbresolver.DBResolver, username string) (bool, *model.User, error) {
	db := dbResolver.GetDB()
	return GetUserByUserNameWithDB(ctx, db, username)
}

func GetUserByUserNameWithDB(ctx context.Context, db *gorm.DB, username string) (bool, *model.User, error) {
	u := model.User{}
	err := db.WithContext(ctx).Model(&u).Where("username = ?", username).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &u, err
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

// InsertUserOperatorLog inserts a new user operation log into the database.
func InsertUserOperatorLog(ctx context.Context, dbResolver *dbresolver.DBResolver, uid string, operator model.UserOperatorType) (*model.UserOperatorLog, error) {
	db := dbResolver.GetDB()
	return InsertUserOperatorLogWithDB(ctx, db, uid, operator)
}

func InsertUserOperatorLogWithDB(ctx context.Context, db *gorm.DB, uid string, operator model.UserOperatorType) (*model.UserOperatorLog, error) {
	creator := token.GetUIDFromCtx(ctx)
	log := model.UserOperatorLog{
		UID:       uid,
		Operator:  operator,
		CreatedAt: time.Now().UnixMilli(),
		Creator:   creator,
	}

	err := db.WithContext(ctx).Create(&log).Error
	return &log, err
}

func InsertUserOperatorLogByModel(ctx context.Context, dbResolver *dbresolver.DBResolver, log *model.UserOperatorLog) error {
	db := dbResolver.GetDB()
	return InsertUserOperatorLogByModelWithDB(ctx, db, log)
}

func InsertUserOperatorLogByModelWithDB(ctx context.Context, db *gorm.DB, log *model.UserOperatorLog) error {
	return db.WithContext(ctx).Create(log).Error
}

// GetUserOperatorLogsByUID retrieves all user operation logs for a specific user by UID.
func GetUserOperatorLogsByUID(ctx context.Context, dbResolver *dbresolver.DBResolver, uid string) ([]model.UserOperatorLog, error) {
	db := dbResolver.GetDB()
	var logs []model.UserOperatorLog
	err := db.WithContext(ctx).Where("uid = ?", uid).Find(&logs).Error
	return logs, err
}

// GetUserOperatorLogByID retrieves a user operation log by its ID.
func GetUserOperatorLogByID(ctx context.Context, dbResolver *dbresolver.DBResolver, id int64) (*model.UserOperatorLog, error) {
	db := dbResolver.GetDB()
	log := model.UserOperatorLog{}
	err := db.WithContext(ctx).Where("id = ?", id).First(&log).Error
	return &log, err
}

// ListUserOperatorLogs retrieves all user operation logs.
func ListUserOperatorLogs(ctx context.Context, dbResolver *dbresolver.DBResolver) ([]model.UserOperatorLog, error) {
	db := dbResolver.GetDB()
	var logs []model.UserOperatorLog
	err := db.WithContext(ctx).Find(&logs).Error
	return logs, err
}
