package model

import "gorm.io/gorm"

// NavGroup represents a navigation menu group stored in the database
type NavGroup struct {
	gorm.Model
	Label     string    `gorm:"uniqueIndex;not null" json:"label"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	Roles     []Role    `gorm:"many2many:nav_group_roles;" json:"roles"`
	Items     []NavItem `gorm:"foreignKey:NavGroupID" json:"items"`
}

// NavItem represents a single navigation menu item stored in the database
type NavItem struct {
	gorm.Model
	NavGroupID uint   `gorm:"not null" json:"nav_group_id"`
	Label      string `gorm:"not null" json:"label"`
	Href       string `gorm:"not null" json:"href"`
	SortOrder  int    `gorm:"default:0" json:"sort_order"`
	Roles      []Role `gorm:"many2many:nav_item_roles;" json:"roles"`
}

// NavGroupResponse is the JSON response for menu config API
type NavGroupResponse struct {
	Label string            `json:"label"`
	Roles []string          `json:"roles"`
	Items []NavItemResponse `json:"items"`
}

// NavItemResponse is the JSON response for a single nav item
type NavItemResponse struct {
	Label string   `json:"label"`
	Href  string   `json:"href"`
	Roles []string `json:"roles"`
}

// ToResponse converts a NavGroup DB model to API response
func (g *NavGroup) ToResponse() NavGroupResponse {
	roles := make([]string, len(g.Roles))
	for i, r := range g.Roles {
		roles[i] = r.Name
	}
	items := make([]NavItemResponse, len(g.Items))
	for i, item := range g.Items {
		items[i] = item.ToResponse()
	}
	return NavGroupResponse{Label: g.Label, Roles: roles, Items: items}
}

// ToResponse converts a NavItem DB model to API response
func (item *NavItem) ToResponse() NavItemResponse {
	roles := make([]string, len(item.Roles))
	for i, r := range item.Roles {
		roles[i] = r.Name
	}
	return NavItemResponse{Label: item.Label, Href: item.Href, Roles: roles}
}

// FilterNavMenuByRoles filters navigation groups and items based on user roles.
// Admin role sees everything.
func FilterNavMenuByRoles(groups []NavGroupResponse, userRoles []string) []NavGroupResponse {
	roleSet := make(map[string]bool, len(userRoles))
	for _, r := range userRoles {
		roleSet[r] = true
	}

	if roleSet[RoleAdmin] {
		return groups
	}

	var result []NavGroupResponse
	for _, group := range groups {
		if len(group.Roles) > 0 && !hasAny(roleSet, group.Roles) {
			continue
		}

		var filteredItems []NavItemResponse
		for _, item := range group.Items {
			if len(item.Roles) == 0 || hasAny(roleSet, item.Roles) {
				filteredItems = append(filteredItems, item)
			}
		}

		if len(filteredItems) > 0 {
			result = append(result, NavGroupResponse{
				Label: group.Label,
				Roles: group.Roles,
				Items: filteredItems,
			})
		}
	}
	return result
}

func hasAny(set map[string]bool, keys []string) bool {
	for _, k := range keys {
		if set[k] {
			return true
		}
	}
	return false
}

// SeedNavMenuConfig defines the default menu structure for seeding
type SeedNavItem struct {
	Label string
	Href  string
	Roles []string
}

type SeedNavGroup struct {
	Label string
	Roles []string
	Items []SeedNavItem
}

var DefaultNavMenuSeed = []SeedNavGroup{
	{
		Label: "Auth",
		Roles: []string{RoleAdmin},
		Items: []SeedNavItem{
			{Label: "กลุ่ม", Href: "/dashboard/groups", Roles: []string{RoleAdmin}},
			{Label: "สิทธิ์", Href: "/dashboard/roles", Roles: []string{RoleAdmin}},
			{Label: "ผู้ใช้", Href: "/dashboard/users", Roles: []string{RoleAdmin}},
		},
	},
	{
		Label: "Catalog",
		Roles: []string{RoleAdmin, RoleMerchant},
		Items: []SeedNavItem{
			{Label: "เมนู", Href: "/dashboard/menus", Roles: []string{RoleAdmin, RoleMerchant}},
			{Label: "วัตถุดิบ (BOM)", Href: "/dashboard/ingredients", Roles: []string{RoleAdmin, RoleMerchant}},
		},
	},
	{
		Label: "Kitchen",
		Roles: []string{RoleAdmin, RoleMerchant},
		Items: []SeedNavItem{
			{Label: "ครัว", Href: "/dashboard/kitchen", Roles: []string{RoleAdmin, RoleMerchant}},
		},
	},
	{
		Label: "Order",
		Roles: []string{},
		Items: []SeedNavItem{
			{Label: "ออเดอร์", Href: "/dashboard/orders", Roles: []string{RoleAdmin, RoleMerchant}},
			{Label: "คิว", Href: "/dashboard/queue", Roles: []string{RoleAdmin, RoleMerchant, RoleRider}},
			{Label: "สั่งอาหาร", Href: "/dashboard/orders/create", Roles: []string{RoleCustomer}},
		},
	},
	{
		Label: "Inventory",
		Roles: []string{RoleAdmin, RoleMerchant},
		Items: []SeedNavItem{
			{Label: "รายการเดินสต๊อก", Href: "/dashboard/inventory", Roles: []string{RoleAdmin, RoleMerchant}},
			{Label: "สต๊อกคงเหลือ", Href: "/dashboard/inventory/stock", Roles: []string{RoleAdmin, RoleMerchant}},
		},
	},
}
