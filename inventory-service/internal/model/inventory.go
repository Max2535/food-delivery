package model

import "time"

const (
	TransactionRestock    = "RESTOCK"
	TransactionDeduction  = "DEDUCTION"
	TransactionAdjustment = "ADJUSTMENT"
)

// RawMaterial is the inventory-owned master list of raw materials.
// CatalogIngredientID links to the ingredient in catalog-service (cross-service reference, not a real FK).
type RawMaterial struct {
	ID                  uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CatalogIngredientID *uint     `gorm:"index" json:"catalog_ingredient_id,omitempty"`
	Name                string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Unit                string    `gorm:"type:varchar(50);not null" json:"unit"` // g, kg, ml, piece
	CurrentStock        float64   `gorm:"type:decimal(12,3);default:0" json:"current_stock"`
	ReorderPoint        float64   `gorm:"type:decimal(12,3);default:0" json:"reorder_point"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// StockTransaction is the immutable audit trail of every stock movement.
type StockTransaction struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RawMaterialID   uint      `gorm:"not null;index" json:"raw_material_id"`
	QuantityChange  float64   `gorm:"type:decimal(12,3);not null" json:"quantity_change"` // positive=restock, negative=deduction
	Type            string    `gorm:"type:varchar(20);not null" json:"type"`
	OrderID         *uint     `json:"order_id,omitempty"`
	CorrelationID   string    `gorm:"type:varchar(100)" json:"correlation_id"`
	Note            string    `gorm:"type:text" json:"note,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}
