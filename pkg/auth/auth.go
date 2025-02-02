package auth

import (
	casbinmodel "github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"gorm.io/gorm"
)

// Adapter represents the Gorm adapter for policy storage.
type Adapter struct {
	db *gorm.DB
}

// NewAdapter creates a new Adapter.
func NewAdapter(db *gorm.DB) *Adapter {
	return &Adapter{db: db}
}

// LoadPolicy loads policy from database.
func (a *Adapter) LoadPolicy(model casbinmodel.Model) error {
	var policies []Permission
	if err := a.db.Find(&policies).Error; err != nil {
		return err
	}

	for _, policy := range policies {
		line := []string{"p", policy.UserID, policy.Resource, policy.Action}
		persist.LoadPolicyArray(line, model)
	}

	return nil
}

// SavePolicy saves policy to database.
func (a *Adapter) SavePolicy(model casbinmodel.Model) error {
	// Clear existing policies
	if err := a.db.Exec("DELETE FROM permissions").Error; err != nil {
		return err
	}

	// Save new policies
	for _, ast := range model["p"] {
		for _, rule := range ast.Policy {
			policy := Permission{
				UserID:   rule[1],
				Resource: rule[2],
				Action:   rule[3],
			}
			if err := a.db.Create(&policy).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// AddPolicy adds a policy rule to the storage.
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	policy := Permission{
		UserID:   rule[1],
		Resource: rule[2],
		Action:   rule[3],
	}
	return a.db.Create(&policy).Error
}

// RemovePolicy removes a policy rule from the storage.
func (a *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return a.db.Where("user_id = ? AND resource = ? AND action = ?", rule[1], rule[2], rule[3]).Delete(&Permission{}).Error
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	query := "1=1"
	var args []interface{}

	if fieldIndex <= 0 && 0 < fieldIndex+len(fieldValues) {
		query += " AND user_id = ?"
		args = append(args, fieldValues[0])
	}
	if fieldIndex <= 1 && 1 < fieldIndex+len(fieldValues) {
		query += " AND resource = ?"
		args = append(args, fieldValues[1])
	}
	if fieldIndex <= 2 && 2 < fieldIndex+len(fieldValues) {
		query += " AND action = ?"
		args = append(args, fieldValues[2])
	}

	return a.db.Where(query, args...).Delete(&Permission{}).Error
}

// Permission represents a permission policy in the database.
type Permission struct {
	ID       uint   `gorm:"primaryKey"`
	UserID   string `gorm:"not null"`
	Resource string `gorm:"not null"`
	Action   string `gorm:"not null"`
}
