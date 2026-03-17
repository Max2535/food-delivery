package repository

import (
	"order-service/internal/model"

	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *model.Order) error
	FindAll() ([]model.Order, error)
	FindByID(id uint) (*model.Order, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) FindAll() ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Preload("Items").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) FindByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Items").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}
