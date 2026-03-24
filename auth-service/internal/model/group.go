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
	Name  string `gorm:"uniqueIndex;not null" json:"name"`
	Roles []Role `gorm:"many2many:group_roles;" json:"roles"`
}

func (g Group) RoleNames() []string {
	names := make([]string, len(g.Roles))
	for i, r := range g.Roles {
		names[i] = r.Name
	}
	return names
}
