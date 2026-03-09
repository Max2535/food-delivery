package service

import (
	"order-service/internal/model"
	"order-service/internal/repository"
)

type OrderService interface {
	CreateOrder(order *model.Order) error
	GetAllOrders() ([]model.Order, error)
	GetOrderByID(id uint) (*model.Order, error)
}

type orderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{repo: repo}
}

func (s *orderService) CreateOrder(order *model.Order) error {
	// Add business logic here (e.g., validation, calculating total)
	if order.Status == "" {
		order.Status = "pending"
	}
	return s.repo.Create(order)
}

func (s *orderService) GetAllOrders() ([]model.Order, error) {
	return s.repo.FindAll()
}

func (s *orderService) GetOrderByID(id uint) (*model.Order, error) {
	return s.repo.FindByID(id)
}
