package model

import "gorm.io/gorm"

type User struct {
	ID        int64      `gorm:"primary_key;AUTO_INCREMENT"`
	UID       string     `gorm:"not null; index:uniq_uid,unique; type:varchar(32)"`
	Username  string     `gorm:"not null; index:idx_username; type:varchar(32)"`
	Role      UserRole   `gorm:"not null"`
	Primary   bool       `gorm:"not null"`
	Tel       string     `gorm:"not null; type:varchar(32)"`
	Email     string     `gorm:"not null; type:varchar(32)"`
	Desc      string     `gorm:"not null; type:varchar(255)"`
	Status    UserStatus `gorm:"not null"`
	CreatedAt int64      `gorm:"autoCreateTime:milli; not null; index:idx_created_at"`
	Creator   string     `gorm:"not null; type:varchar(32)"`
	UpdatedAt int64      `gorm:"autoUpdateTime:milli; not null"`
	Updater   string     `gorm:"not null; type:varchar(32)"`
	gorm.DeletedAt
}
type UserRole string

const (
	Admin  UserRole = "admin"
	Normal UserRole = "normal"
)

type UserStatus string

const (
	PrimaryAccountID = 1

	UserStatusEnabled  UserStatus = "enabled"
	UserStatusDisabled UserStatus = "disabled"
	UserStatusLocked   UserStatus = "locked"
)

func (User) TableName() string {
	return "users"
}

var DefaultUser = User{
	Role:     Normal,
	Status:   UserStatusEnabled,
	Primary:  true,
	UID:      "",
	Username: "",
	Tel:      "",
	Email:    "",
	Desc:     "",
}

type UserOperatorLog struct {
	ID        int64  `gorm:"primary_key;AUTO_INCREMENT"`
	UID       string `gorm:"not null; index:uid"`
	Operator  string `gorm:"not null; type:varchar(255)"`
	CreatedAt int64  `gorm:"autoCreateTime:milli; not null; index:idx_created_at"`
	Creator   string `gorm:"not null; type:varchar(32)"`
}
