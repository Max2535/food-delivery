package model

import (
	"gorm.io/gorm"
)

const (
	New       = "new"
	Pending   = "pending"
	Confirmed = "confirmed"
	Shipped   = "shipped"
	Failed    = "failed"
	Delivered = "delivered"
	Finished  = "finished"
)

type Order struct {
	gorm.Model
	CustomerID  string  `gorm:"type:varchar(100);not null" json:"customer_id"`
	TotalAmount float64 `gorm:"type:decimal(10,2)" json:"total_amount"`

	// Status => new, pending, confirmed, shipped, failed, delivered, finished
	Status          string      `gorm:"type:varchar(20);default:'pending'" json:"status"`
	DeliveryAddress string      `gorm:"type:text" json:"delivery_address"`
	Items           []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

type OrderItem struct {
	gorm.Model
	OrderID    uint    `gorm:"not null" json:"order_id"`
	MenuItemID uint    `gorm:"not null" json:"menu_item_id"`
	Quantity   int     `gorm:"not null" json:"quantity"`
	UnitPrice  float64 `gorm:"type:decimal(10,2);not null" json:"unit_price"`
}
