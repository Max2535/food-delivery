package repository

import (
	"kitchen-service/internal/model"
	"gorm.io/gorm"
)

type KitchenRepository interface {
	Create(ticket *model.KitchenTicket) error
	UpdateStatus(orderID uint, status string) error
	GetByOrderID(orderID uint) (*model.KitchenTicket, error)
	GetQueue() ([]*model.KitchenTicket, error)
}

type kitchenRepository struct {
	db *gorm.DB
}

func NewKitchenRepository(db *gorm.DB) KitchenRepository {
	return &kitchenRepository{db: db}
}

func (r *kitchenRepository) Create(ticket *model.KitchenTicket) error {
	return r.db.Create(ticket).Error
}

func (r *kitchenRepository) UpdateStatus(orderID uint, status string) error {
	// อัปเดตสถานะโดยอ้างอิงจาก OrderID ที่ได้รับมาจาก Order Service
	return r.db.Model(&model.KitchenTicket{}).
		Where("order_id = ?", orderID).
		Update("status", status).Error
}

func (r *kitchenRepository) GetByOrderID(orderID uint) (*model.KitchenTicket, error) {
	var ticket model.KitchenTicket
	err := r.db.Where("order_id = ?", orderID).First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *kitchenRepository) GetQueue() ([]*model.KitchenTicket, error) {
	var tickets []*model.KitchenTicket
	// Exclude Served status, order by Priority ASC (Urgent=1), then First-In-First-Out
	err := r.db.Where("status != ?", model.StatusServed).
		Order("priority ASC").
		Order("created_at ASC").
		Find(&tickets).Error
	return tickets, err
}
