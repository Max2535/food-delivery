package model

import "time"

const (
	PriorityUrgent = 1
	PriorityNormal = 2
	PriorityLow    = 3
)

const (
	StatusPending      = "Pending"
	StatusPreparing    = "Preparing"
	StatusReadyToServe = "Ready to Serve"
	StatusServed       = "Served"
)

type KitchenTicket struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   uint      `gorm:"uniqueIndex" json:"order_id"`
	Items     string    `json:"items"`    // รายการอาหาร (อาจจะเก็บเป็น JSON string)
	Status    string    `json:"status"`   // "Pending", "Preparing", "Ready to Serve", "Served"
	Priority  int       `json:"priority"` // 1: Urgent, 2: Normal, 3: Low
	CreatedAt time.Time `json:"created_at"`
}
