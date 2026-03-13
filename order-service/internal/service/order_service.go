package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"order-service/internal/model"
	"order-service/internal/repository"

	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"github.com/sony/gobreaker"
)

var cb *gobreaker.CircuitBreaker

func init() {
	settings := gobreaker.Settings{
		Name:        "Kitchen-Service",
		MaxRequests: 3,
		Interval:    5 * time.Second,
		Timeout:     30 * time.Second, // ระยะเวลาที่วงจรจะเปิดค้างไว้ก่อนลองกลับมาต่อใหม่
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// ถ้าพังติดต่อกัน 3 ครั้ง ให้ตัดวงจรทันที
			return counts.ConsecutiveFailures > 3
		},
	}
	cb = gobreaker.NewCircuitBreaker(settings)
}

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
	if err != nil {
		return err
	}

	// หุ้มส่วนที่คุยกับ Kitchen ด้วย Circuit Breaker
	_, err = cb.Execute(func() (interface{}, error) {
		// โค้ดที่ยิง HTTP หรือ RabbitMQ ไปหา Kitchen
		return nil, s.publishToKitchen(order)
	})

	if err != nil {
		// Newman แนะนำ: ถ้าวงจรตัด ให้หาทางออกสำรอง (Fallback)
		// เช่น บันทึกไว้ในคิวสำรอง หรือบอกลูกค้าว่า "คิวครัวเต็ม" แทนที่จะปล่อยให้หมุนค้าง
		return fmt.Errorf("Kitchen service is unavailable: %v", err)
	}

	return nil
}

func (s *orderService) publishToKitchen(order *model.Order) error {
	// 1. เชื่อมต่อ RabbitMQ
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// 2. ประกาศ Exchange (ควรใช้ชื่อที่สื่อถึง Domain)
	err = ch.ExchangeDeclare("order_events", "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	// 3. เตรียมข้อมูล JSON
	body, _ := json.Marshal(map[string]interface{}{
		"order_id": order.ID,
		"items":    "[]", // TODO: Implement Order Items in Model
	})

	correlationID := uuid.New().String()
	log.Info().
		Str("service", "order-service").
		Uint("order_id", order.ID).
		Str("correlation_id", correlationID).
		Msg("Publishing event for order")

	// 4. Publish Message พร้อม Routing Key
	return ch.PublishWithContext(context.Background(),
		"order_events",       // exchange
		"order.created",      // routing key
		false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			Body:          body,
		})
}

func (s *orderService) GetAllOrders() ([]model.Order, error) {
	return s.repo.FindAll()
}

func (s *orderService) GetOrderByID(id uint) (*model.Order, error) {
	return s.repo.FindByID(id)
}

func getKitchenServiceAddress() string {
    config := consul.DefaultConfig()
    // Config consul address if needed, default is localhost:8500
    if consulAddr := os.Getenv("CONSUL_ADDRESS"); consulAddr != "" {
        config.Address = consulAddr
    }
    
    client, err := consul.NewClient(config)
    if err != nil {
        log.Error().Err(err).Msg("Error creating consul client")
        return ""
    }

    // ถามหาบริการที่ชื่อ kitchen-service
    services, _, err := client.Health().Service("kitchen-service", "", true, nil)
    if err != nil {
        log.Error().Err(err).Msg("Error discovering kitchen-service")
        return ""
    }
    
    if len(services) > 0 {
        addr := services[0].Service.Address
        port := services[0].Service.Port
        address := fmt.Sprintf("http://%s:%d", addr, port)
        log.Info().Str("service", "order-service").Str("address", address).Msg("Discovered kitchen-service")
        return address
    }
    
    log.Warn().Str("service", "order-service").Msg("kitchen-service not found in Consul")
    return ""
}

