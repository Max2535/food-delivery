package service

import (
	"context"
	"encoding/json"
	"os"

	"order-service/internal/model"
	"order-service/internal/repository"

	amqp "github.com/rabbitmq/amqp091-go"
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

	err := s.repo.Create(order)
	if err == nil {
		s.publishToKitchen(order)
	}
	return err
}

func (s *orderService) publishToKitchen(order *model.Order) {
	// 1. เชื่อมต่อ RabbitMQ
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return
	}
	defer ch.Close()

	// 2. ประกาศ Exchange (ควรใช้ชื่อที่สื่อถึง Domain)
	ch.ExchangeDeclare("order_events", "topic", true, false, false, false, nil)

	// 3. เตรียมข้อมูล JSON
	body, _ := json.Marshal(map[string]interface{}{
		"order_id": order.ID,
		"items":    "[]", // TODO: Implement Order Items in Model
	})

	// 4. Publish Message พร้อม Routing Key
	ch.PublishWithContext(context.Background(),
		"order_events",       // exchange
		"order.created",      // routing key
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (s *orderService) GetAllOrders() ([]model.Order, error) {
	return s.repo.FindAll()
}

func (s *orderService) GetOrderByID(id uint) (*model.Order, error) {
	return s.repo.FindByID(id)
}
