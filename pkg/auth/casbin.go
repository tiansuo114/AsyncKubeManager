package auth

import (
	"github.com/casbin/casbin/v2"
	"gorm.io/gorm"
)

// Enforcer represents the Casbin enforcer with database storage.
type Enforcer struct {
	e *casbin.Enforcer
}

// NewEnforcer creates a new Enforcer.
func NewEnforcer(db *gorm.DB) (*Enforcer, error) {
	adapter := NewAdapter(db)
	e, err := casbin.NewEnforcer("configs/casbin_model.conf", adapter)
	if err != nil {
		return nil, err
	}
	return &Enforcer{e: e}, nil
}

// AddPolicy adds a policy rule.
func (e *Enforcer) AddPolicy(userID, resource, action string) error {
	_, err := e.e.AddPolicy(userID, resource, action)
	return err
}

// RemovePolicy removes a policy rule.
func (e *Enforcer) RemovePolicy(userID, resource, action string) error {
	_, err := e.e.RemovePolicy(userID, resource, action)
	return err
}

// Enforce checks if a user has permission to perform an action on a resource.
func (e *Enforcer) Enforce(userID, resource, action string) (bool, error) {
	return e.e.Enforce(userID, resource, action)
}
