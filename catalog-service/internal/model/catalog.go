package model

import "time"

// MenuItem represents a food item in the catalog
type MenuItem struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Category    string    `gorm:"type:varchar(100);not null" json:"category"`
	IsAvailable bool      `gorm:"default:true" json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	BOM         []BOMItem `gorm:"foreignKey:MenuItemID" json:"bom,omitempty"`
}

// Ingredient represents a raw material used in recipes
type Ingredient struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Unit      string    `gorm:"type:varchar(50);not null" json:"unit"` // e.g. "g", "ml", "piece"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BOMItem (Bill of Materials) links a MenuItem to either a raw Ingredient or another MenuItem
// acting as a sub-recipe. Exactly one of IngredientID or SubMenuItemID must be non-nil.
type BOMItem struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	MenuItemID    uint      `gorm:"not null;index" json:"menu_item_id"`
	IngredientID  *uint     `json:"ingredient_id,omitempty"`
	SubMenuItemID *uint     `json:"sub_menu_item_id,omitempty"`
	Quantity      float64   `gorm:"type:decimal(10,3);not null" json:"quantity"`
	Ingredient    *Ingredient `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
	SubMenuItem   *MenuItem   `gorm:"foreignKey:SubMenuItemID" json:"sub_menu_item,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// KitchenStation represents a cooking station (e.g. Hot Kitchen, Cold Kitchen, Bar)
type KitchenStation struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MenuStationMapping is the join table that maps a MenuItem to its KitchenStation
type MenuStationMapping struct {
	MenuItemID       uint `gorm:"primaryKey" json:"menu_item_id"`
	KitchenStationID uint `gorm:"primaryKey" json:"kitchen_station_id"`
}

// ── Case 1: Customer Choice (e.g. เลือกเส้น) ─────────────────────────────────

// BOMChoiceGroup defines a group of mutually-exclusive ingredient options for a menu item
// e.g. "เลือกเส้น" for ก๋วยเตี๋ยว
type BOMChoiceGroup struct {
	ID         uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	MenuItemID uint              `gorm:"not null;index" json:"menu_item_id"`
	Name       string            `gorm:"type:varchar(100);not null" json:"name"`       // e.g. "เลือกเส้น"
	IsRequired bool              `gorm:"default:true" json:"is_required"`               // must pick one
	MinChoices int               `gorm:"default:1" json:"min_choices"`
	MaxChoices int               `gorm:"default:1" json:"max_choices"`
	Options    []BOMChoiceOption `gorm:"foreignKey:GroupID" json:"options,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// BOMChoiceOption is one selectable option inside a BOMChoiceGroup
type BOMChoiceOption struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupID      uint       `gorm:"not null;index" json:"group_id"`
	IngredientID uint       `gorm:"not null" json:"ingredient_id"`
	Quantity     float64    `gorm:"type:decimal(10,3);not null" json:"quantity"`
	ExtraPrice   float64    `gorm:"type:decimal(10,2);default:0" json:"extra_price"`
	Ingredient   Ingredient `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ── Case 2: Add-on (e.g. เพิ่มไข่ดาว) ────────────────────────────────────────

// MenuAddOn represents an optional extra that a customer can add to a menu item
// e.g. ไข่ดาว +10฿ for กะเพรา
type MenuAddOn struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	MenuItemID   uint       `gorm:"not null;index" json:"menu_item_id"`
	IngredientID uint       `gorm:"not null" json:"ingredient_id"`
	Quantity     float64    `gorm:"type:decimal(10,3);not null" json:"quantity"`
	ExtraPrice   float64    `gorm:"type:decimal(10,2);default:0" json:"extra_price"`
	IsAvailable  bool       `gorm:"default:true" json:"is_available"`
	Ingredient   Ingredient `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ── Case 3: Portion Size (e.g. ธรรมดา / พิเศษ) ───────────────────────────────

// MenuPortionSize defines size variants for a menu item
// e.g. "พิเศษ" multiplies all BOM quantities by 1.5 and adds 15฿
type MenuPortionSize struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	MenuItemID         uint      `gorm:"not null;index" json:"menu_item_id"`
	Name               string    `gorm:"type:varchar(100);not null" json:"name"`           // e.g. "ธรรมดา", "พิเศษ"
	QuantityMultiplier float64   `gorm:"type:decimal(5,2);default:1.0" json:"quantity_multiplier"`
	ExtraPrice         float64   `gorm:"type:decimal(10,2);default:0" json:"extra_price"`
	IsDefault          bool      `gorm:"default:false" json:"is_default"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
