package model

import "gorm.io/gorm"

const (
	GroupUser     = "user"
	GroupAdmin    = "admin"
	GroupRider    = "rider"
	GroupMerchant = "merchant"
	GroupCustomer = "customer"
)

type Group struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null" json:"name"`
	Description string `json:"description"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	Roles       []Role `gorm:"many2many:group_roles;" json:"roles"`
	Users       []User `gorm:"foreignKey:GroupID" json:"users,omitempty"`
}

func (g Group) RoleNames() []string {
	names := make([]string, len(g.Roles))
	for i, r := range g.Roles {
		names[i] = r.Name
	}
	return names
}
