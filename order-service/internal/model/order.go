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

	// GORM จะจัดการค่า CreatedAt และ UpdatedAt ให้โดยอัตโนมัติ
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
