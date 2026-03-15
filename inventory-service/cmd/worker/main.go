package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"inventory-service/internal/catalog"
	"inventory-service/internal/model"
	"inventory-service/internal/repository"
	"inventory-service/internal/service"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// kitchenTicketEvent matches the payload published by kitchen-service worker
type kitchenTicketEvent struct {
	OrderID  uint   `json:"order_id"`
	TicketID uint   `json:"ticket_id"`
	Items    string `json:"items"` // JSON array of {menu_item_id, quantity, portion_multiplier}
}

type orderItem struct {
	MenuItemID        uint    `json:"menu_item_id"`
	Quantity          int     `json:"quantity"`
	PortionMultiplier float64 `json:"portion_multiplier"`
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Warn().Msg("Warning: .env file not found")
	}

	// Database
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal().Msg("DB_URL is not set")
	}

	var db *gorm.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
		if err == nil {
			break
		}
		log.Warn().Err(err).Int("attempt", i+1).Msg("Failed to connect to database")
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to database after retries")
	}

	db.AutoMigrate(&model.RawMaterial{}, &model.StockTransaction{})

	// Catalog client
	catalogURL := os.Getenv("CATALOG_SERVICE_URL")
	if catalogURL == "" {
		catalogURL = "http://catalog-service-api:3003"
	}
	catalogClient := catalog.NewClient(catalogURL)

	materialRepo := repository.NewMaterialRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	stockSvc := service.NewStockService(materialRepo, transactionRepo, catalogClient)

	// RabbitMQ
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	var conn *amqp.Connection
	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			break
		}
		log.Warn().Err(err).Int("attempt", i+1).Msg("Failed to connect to RabbitMQ")
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to RabbitMQ")
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open channel")
	}
	defer ch.Close()

	// Declare kitchen_events exchange (published by kitchen-service worker)
	if err := ch.ExchangeDeclare("kitchen_events", "topic", true, false, false, false, nil); err != nil {
		log.Fatal().Err(err).Msg("Failed to declare kitchen_events exchange")
	}

	q, err := ch.QueueDeclare("inventory_kitchen_queue", true, false, false, false, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to declare queue")
	}

	if err := ch.QueueBind(q.Name, "kitchen.ticket_created", "kitchen_events", false, nil); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind queue")
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register consumer")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Info().Str("service", "inventory-worker").Msg("Shutting down...")
		cancel()
	}()

	log.Info().Str("service", "inventory-worker").Msg("[*] Waiting for kitchen.ticket_created events")

	for {
		select {
		case <-ctx.Done():
			return
		case d := <-msgs:
			correlationID := d.CorrelationId
			if correlationID == "" {
				correlationID = "unknown"
			}

			var event kitchenTicketEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Error().Err(err).Str("service", "inventory-worker").Msg("Failed to unmarshal event")
				continue
			}

			log.Info().
				Str("service", "inventory-worker").
				Str("correlation_id", correlationID).
				Uint("order_id", event.OrderID).
				Uint("ticket_id", event.TicketID).
				Msg("Received kitchen.ticket_created — processing stock deduction")

			// Parse items from the event
			var items []orderItem
			if err := json.Unmarshal([]byte(event.Items), &items); err != nil || len(items) == 0 {
				log.Warn().
					Str("service", "inventory-worker").
					Str("correlation_id", correlationID).
					Uint("order_id", event.OrderID).
					Msg("Items field is empty or unparseable — skipping auto-deduction (waiting for Order Service to include item data)")
				continue
			}

			// Convert to DeductItem slice
			deductItems := make([]service.DeductItem, len(items))
			for i, it := range items {
				mult := it.PortionMultiplier
				if mult <= 0 {
					mult = 1.0
				}
				deductItems[i] = service.DeductItem{
					MenuItemID:        it.MenuItemID,
					Quantity:          it.Quantity,
					PortionMultiplier: mult,
				}
			}

			if err := stockSvc.DeductByBOM(&event.OrderID, deductItems, correlationID); err != nil {
				log.Error().Err(err).
					Str("service", "inventory-worker").
					Str("correlation_id", correlationID).
					Uint("order_id", event.OrderID).
					Msg("Stock deduction failed")
			} else {
				log.Info().
					Str("service", "inventory-worker").
					Str("correlation_id", correlationID).
					Uint("order_id", event.OrderID).
					Msg("Stock deduction completed")
			}
		}
	}
}
