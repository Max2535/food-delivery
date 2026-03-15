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
}

// Ingredient represents a raw material used in recipes
type Ingredient struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Unit      string    `gorm:"type:varchar(50);not null" json:"unit"` // e.g. "g", "ml", "piece"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BOMItem (Bill of Materials) links a MenuItem to its required Ingredients with quantity
type BOMItem struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	MenuItemID   uint       `gorm:"not null;index" json:"menu_item_id"`
	IngredientID uint       `gorm:"not null" json:"ingredient_id"`
	Quantity     float64    `gorm:"type:decimal(10,3);not null" json:"quantity"`
	Ingredient   Ingredient `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
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
