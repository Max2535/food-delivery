package model

import "gorm.io/gorm"

// Permission constants — ใช้ format "module.resource.action"
const (
	// Auth
	PermAuthGroupsView   = "auth.groups.view"
	PermAuthGroupsManage = "auth.groups.manage"
	PermAuthRolesView    = "auth.roles.view"
	PermAuthRolesManage  = "auth.roles.manage"
	PermAuthUsersView    = "auth.users.view"
	PermAuthUsersManage  = "auth.users.manage"

	// Catalog
	PermCatalogMenusView         = "catalog.menus.view"
	PermCatalogMenusManage       = "catalog.menus.manage"
	PermCatalogIngredientsView   = "catalog.ingredients.view"
	PermCatalogIngredientsManage = "catalog.ingredients.manage"
	PermCatalogStationsView      = "catalog.stations.view"
	PermCatalogStationsManage    = "catalog.stations.manage"

	// Kitchen
	PermKitchenView   = "kitchen.view"
	PermKitchenManage = "kitchen.manage"

	// Order
	PermOrdersView   = "orders.view"
	PermOrdersManage = "orders.manage"
	PermOrdersCreate = "orders.create"
	PermQueueView    = "orders.queue.view"

	// Inventory
	PermInventoryView   = "inventory.view"
	PermInventoryManage = "inventory.manage"
)

type Permission struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null" json:"name"`
	Description string `json:"description"`
}

// AllPermissions lists every permission for seeding
var AllPermissions = []struct {
	Name        string
	Description string
}{
	{PermAuthGroupsView, "ดูรายการกลุ่ม"},
	{PermAuthGroupsManage, "จัดการกลุ่ม (เพิ่ม/แก้ไข/ลบ)"},
	{PermAuthRolesView, "ดูรายการสิทธิ์"},
	{PermAuthRolesManage, "จัดการสิทธิ์"},
	{PermAuthUsersView, "ดูรายการผู้ใช้"},
	{PermAuthUsersManage, "จัดการผู้ใช้"},
	{PermCatalogMenusView, "ดูเมนูอาหาร"},
	{PermCatalogMenusManage, "จัดการเมนูอาหาร"},
	{PermCatalogIngredientsView, "ดูวัตถุดิบ"},
	{PermCatalogIngredientsManage, "จัดการวัตถุดิบ"},
	{PermKitchenView, "ดูสถานะครัว"},
	{PermKitchenManage, "จัดการครัว"},
	{PermOrdersView, "ดูออเดอร์"},
	{PermOrdersManage, "จัดการออเดอร์"},
	{PermOrdersCreate, "สร้างออเดอร์ (สั่งอาหาร)"},
	{PermQueueView, "ดูคิว"},
	{PermInventoryView, "ดูสต๊อก"},
	{PermInventoryManage, "จัดการสต๊อก"},
}

// RolePermissions maps each role to its default permissions
var RolePermissions = map[string][]string{
	RoleAdmin: {
		PermAuthGroupsView, PermAuthGroupsManage,
		PermAuthRolesView, PermAuthRolesManage,
		PermAuthUsersView, PermAuthUsersManage,
		PermCatalogMenusView, PermCatalogMenusManage,
		PermCatalogIngredientsView, PermCatalogIngredientsManage,
		PermCatalogStationsView, PermCatalogStationsManage,
		PermKitchenView, PermKitchenManage,
		PermOrdersView, PermOrdersManage, PermOrdersCreate, PermQueueView,
		PermInventoryView, PermInventoryManage,
	},
	RoleMerchant: {
		PermCatalogMenusView, PermCatalogMenusManage,
		PermCatalogIngredientsView, PermCatalogIngredientsManage,
		PermCatalogStationsView, PermCatalogStationsManage,
		PermKitchenView, PermKitchenManage,
		PermOrdersView, PermOrdersManage, PermQueueView,
		PermInventoryView, PermInventoryManage,
	},
	RoleRider: {
		PermQueueView,
		PermOrdersView,
	},
	RoleCustomer: {
		PermOrdersCreate,
		PermOrdersView,
	},
	RoleUser: {},
}
