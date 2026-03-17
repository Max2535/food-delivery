package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"order-service/internal/model"
	"order-service/internal/repository"

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
	CreateOrder(order *model.Order, correlationID string) error
	CreateOrderFromRequest(req *model.CreateOrderRequest, correlationID string) (*model.Order, error)
	GetAllOrders() ([]model.Order, error)
	GetOrderByID(id uint) (*model.Order, error)
}

type orderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{repo: repo}
}

// CreateOrderFromRequest สร้าง Order จาก DTO โดยคำนวณราคาแต่ละ item และราคารวม
func (s *orderService) CreateOrderFromRequest(req *model.CreateOrderRequest, correlationID string) (*model.Order, error) {
	// Validate
	if req.CustomerID == "" {
		return nil, fmt.Errorf("customer_id is required")
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("at least one item is required")
	}

	// Build Order with Items
	order := &model.Order{
		CustomerID: req.CustomerID,
		Status:     "pending",
	}

	var totalAmount float64
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			item.Quantity = 1
		}
		itemTotal := item.UnitPrice * float64(item.Quantity)
		totalAmount += itemTotal

		order.Items = append(order.Items, model.OrderItem{
			MenuItemID:   item.MenuItemID,
			MenuItemName: item.MenuItemName,
			UnitPrice:    item.UnitPrice,
			Quantity:     item.Quantity,
			TotalPrice:   itemTotal,
		})
	}
	order.TotalAmount = totalAmount

	// Persist (GORM จะ create ทั้ง Order + Items ใน transaction เดียว)
	if err := s.repo.Create(order); err != nil {
		return nil, err
	}

	// หุ้มส่วนที่คุยกับ Kitchen ด้วย Circuit Breaker
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, s.publishToKitchen(order, correlationID)
	})

	if err != nil {
		log.Warn().Err(err).Uint("order_id", order.ID).Msg("Failed to publish to kitchen (order still saved)")
	}

	return order, nil
}

func (s *orderService) CreateOrder(order *model.Order, correlationID string) error {
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
		return nil, s.publishToKitchen(order, correlationID)
	})

	if err != nil {
		// Newman แนะนำ: ถ้าวงจรตัด ให้หาทางออกสำรอง (Fallback)
		return fmt.Errorf("Kitchen service is unavailable: %v", err)
	}

	return nil
}

func (s *orderService) publishToKitchen(order *model.Order, correlationID string) error {
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

	// 3. เตรียมข้อมูล JSON พร้อม items จริง
	type publishItem struct {
		MenuItemID   uint    `json:"menu_item_id"`
		MenuItemName string  `json:"menu_item_name"`
		Quantity     int     `json:"quantity"`
		UnitPrice    float64 `json:"unit_price"`
		TotalPrice   float64 `json:"total_price"`
	}

	var items []publishItem
	for _, item := range order.Items {
		items = append(items, publishItem{
			MenuItemID:   item.MenuItemID,
			MenuItemName: item.MenuItemName,
			Quantity:     item.Quantity,
			UnitPrice:    item.UnitPrice,
			TotalPrice:   item.TotalPrice,
		})
	}

	body, _ := json.Marshal(map[string]interface{}{
		"order_id":     order.ID,
		"customer_id":  order.CustomerID,
		"total_amount": order.TotalAmount,
		"items":        items,
	})

	// ใช้ correlationID ที่รับมาจาก Gateway แทนการสร้างใหม่
	log.Info().
		Str("service", "order-service").
		Uint("order_id", order.ID).
		Str("correlation_id", correlationID).
		Int("item_count", len(items)).
		Msg("Publishing event for order")

	// 4. Publish Message พร้อม Routing Key
	return ch.PublishWithContext(context.Background(),
		"order_events",       // exchange
		"order.created",      // routing key
		false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID, // ✅ ใช้ ID เดิมจาก Gateway
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
