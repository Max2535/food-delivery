package model

import (
	"time"
)

type Order struct {
	// ใช้ uint เป็น ID ตามมาตรฐาน GORM พร้อมกำหนด auto-increment
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// CustomerID ควรเป็น string (UUID) เพื่อรองรับการเชื่อมโยงกับ Identity Service อื่น
	CustomerID string `gorm:"type:varchar(100);not null" json:"customer_id"`

	TotalAmount float64 `gorm:"type:decimal(10,2)" json:"total_amount"`

	// Status สำหรับทำ State Machine เช่น Pending, Confirmed, Shipped
	Status string `gorm:"type:varchar(20);default:'Pending'" json:"status"`

	// Items — รายการสินค้าในออเดอร์
	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`

	// GORM จะจัดการค่า CreatedAt และ UpdatedAt ให้โดยอัตโนมัติ
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrderItem เก็บข้อมูลแต่ละรายการในออเดอร์
// ใช้ Ref + Snapshot: เก็บทั้ง MenuItemID (trace กลับ) และ snapshot ชื่อ/ราคา ณ ตอนสั่ง
type OrderItem struct {
	ID      uint `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID uint `gorm:"not null;index" json:"order_id"`

	// Reference — เพื่อ trace กลับได้ว่ามาจาก menu ไหน
	MenuItemID uint `gorm:"not null" json:"menu_item_id"`

	// Snapshot — เก็บข้อมูล ณ ตอนสั่ง ป้องกันกรณี menu เปลี่ยนชื่อ/ราคา/ถูกลบ
	MenuItemName string  `gorm:"type:varchar(255);not null" json:"menu_item_name"`
	UnitPrice    float64 `gorm:"type:decimal(10,2);not null" json:"unit_price"`

	Quantity   int     `gorm:"not null;default:1" json:"quantity"`
	TotalPrice float64 `gorm:"type:decimal(10,2);not null" json:"total_price"`
}

// ── DTOs for Create Order Request ─────────────────────────────────────────────

// CreateOrderRequest เป็น DTO สำหรับรับข้อมูลจาก client
type CreateOrderRequest struct {
	CustomerID string                  `json:"customer_id"`
	Items      []CreateOrderItemRequest `json:"items"`
}

// CreateOrderItemRequest เป็น DTO สำหรับแต่ละ item ใน request
type CreateOrderItemRequest struct {
	MenuItemID   uint    `json:"menu_item_id"`
	MenuItemName string  `json:"menu_item_name"`
	UnitPrice    float64 `json:"unit_price"`
	Quantity     int     `json:"quantity"`
}
