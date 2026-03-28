package model

import "gorm.io/gorm"

// NavGroup represents a navigation menu group stored in the database
type NavGroup struct {
	gorm.Model
	Label       string       `gorm:"uniqueIndex;not null" json:"label"`
	SortOrder   int          `gorm:"default:0" json:"sort_order"`
	Permissions []Permission `gorm:"many2many:nav_group_permissions;" json:"permissions"`
	Items       []NavItem    `gorm:"foreignKey:NavGroupID" json:"items"`
}

// NavItem represents a single navigation menu item stored in the database
type NavItem struct {
	gorm.Model
	NavGroupID  uint         `gorm:"not null" json:"nav_group_id"`
	Label       string       `gorm:"not null" json:"label"`
	Href        string       `gorm:"not null" json:"href"`
	SortOrder   int          `gorm:"default:0" json:"sort_order"`
	Permissions []Permission `gorm:"many2many:nav_item_permissions;" json:"permissions"`
}

// ── API Response ─────────────────────────────────────────────────

type NavGroupResponse struct {
	Label       string            `json:"label"`
	Permissions []string          `json:"permissions"`
	Items       []NavItemResponse `json:"items"`
}

type NavItemResponse struct {
	Label       string   `json:"label"`
	Href        string   `json:"href"`
	Permissions []string `json:"permissions"`
}

func (g *NavGroup) ToResponse() NavGroupResponse {
	perms := make([]string, len(g.Permissions))
	for i, p := range g.Permissions {
		perms[i] = p.Name
	}
	items := make([]NavItemResponse, len(g.Items))
	for i, item := range g.Items {
		items[i] = item.ToResponse()
	}
	return NavGroupResponse{Label: g.Label, Permissions: perms, Items: items}
}

func (item *NavItem) ToResponse() NavItemResponse {
	perms := make([]string, len(item.Permissions))
	for i, p := range item.Permissions {
		perms[i] = p.Name
	}
	return NavItemResponse{Label: item.Label, Href: item.Href, Permissions: perms}
}

// ── Filter ───────────────────────────────────────────────────────

// FilterNavMenuByPermissions filters groups/items by user's collected permissions.
func FilterNavMenuByPermissions(groups []NavGroupResponse, userPermissions []string) []NavGroupResponse {
	permSet := make(map[string]bool, len(userPermissions))
	for _, p := range userPermissions {
		permSet[p] = true
	}

	var result []NavGroupResponse
	for _, group := range groups {
		if len(group.Permissions) > 0 && !hasAny(permSet, group.Permissions) {
			continue
		}

		var filteredItems []NavItemResponse
		for _, item := range group.Items {
			if len(item.Permissions) == 0 || hasAny(permSet, item.Permissions) {
				filteredItems = append(filteredItems, item)
			}
		}

		if len(filteredItems) > 0 {
			result = append(result, NavGroupResponse{
				Label:       group.Label,
				Permissions: group.Permissions,
				Items:       filteredItems,
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

// ── Seed Data ────────────────────────────────────────────────────

type SeedNavItem struct {
	Label       string
	Href        string
	Permissions []string
}

type SeedNavGroup struct {
	Label       string
	Permissions []string
	Items       []SeedNavItem
}

var DefaultNavMenuSeed = []SeedNavGroup{
	{
		Label:       "Auth",
		Permissions: []string{PermAuthGroupsView, PermAuthRolesView, PermAuthUsersView},
		Items: []SeedNavItem{
			{Label: "กลุ่ม", Href: "/dashboard/groups", Permissions: []string{PermAuthGroupsView}},
			{Label: "สิทธิ์", Href: "/dashboard/roles", Permissions: []string{PermAuthRolesView}},
			{Label: "ผู้ใช้", Href: "/dashboard/users", Permissions: []string{PermAuthUsersView}},
		},
	},
	{
		Label:       "Catalog",
		Permissions: []string{PermCatalogMenusView, PermCatalogIngredientsView},
		Items: []SeedNavItem{
			{Label: "เมนู", Href: "/dashboard/menus", Permissions: []string{PermCatalogMenusView}},
			{Label: "วัตถุดิบ (BOM)", Href: "/dashboard/ingredients", Permissions: []string{PermCatalogIngredientsView}},
		},
	},
	{
		Label:       "Kitchen",
		Permissions: []string{PermKitchenView},
		Items: []SeedNavItem{
			{Label: "ครัว", Href: "/dashboard/kitchen", Permissions: []string{PermKitchenView}},
		},
	},
	{
		Label:       "Order",
		Permissions: []string{PermOrdersView, PermOrdersCreate, PermQueueView},
		Items: []SeedNavItem{
			{Label: "ออเดอร์", Href: "/dashboard/orders", Permissions: []string{PermOrdersView}},
			{Label: "คิว", Href: "/dashboard/queue", Permissions: []string{PermQueueView}},
			{Label: "สั่งอาหาร", Href: "/dashboard/orders/create", Permissions: []string{PermOrdersCreate}},
		},
	},
	{
		Label:       "Inventory",
		Permissions: []string{PermInventoryView},
		Items: []SeedNavItem{
			{Label: "รายการเดินสต๊อก", Href: "/dashboard/inventory", Permissions: []string{PermInventoryView}},
			{Label: "สต๊อกคงเหลือ", Href: "/dashboard/inventory/stock", Permissions: []string{PermInventoryView}},
		},
	},
}
