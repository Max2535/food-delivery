package model

import (
	"time"
)

type KitchenTicket struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   uint      `gorm:"uniqueIndex" json:"order_id"`
	Items     string    `gorm:"type:text" json:"items"` // รายการอาหาร (เก็บเป็น JSON string)
	Status    string    `json:"status"`               // "Received", "Cooking", "Ready"
	CreatedAt time.Time `json:"created_at"`
}
