package model

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	Name        string       `gorm:"uniqueIndex;not null" json:"name"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

func (r Role) PermissionNames() []string {
	names := make([]string, len(r.Permissions))
	for i, p := range r.Permissions {
		names[i] = p.Name
	}
	return names
}
