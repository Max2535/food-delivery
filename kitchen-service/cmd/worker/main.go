package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"kitchen-service/internal/model"
	"kitchen-service/internal/repository"
	"kitchen-service/internal/service"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 1. Database Connection (Kitchen DB)
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		log.Fatal("DB_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Kitchen DB: %v", err)
	}

	// Auto Migrate
	db.AutoMigrate(&model.KitchenTicket{})

	// 2. Setup Layers
	repo := repository.NewKitchenRepository(db)
	kitchenSvc := service.NewKitchenService(repo)

	// 3. Setup RabbitMQ Connection
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// 4. Setup Exchange & Queue
	err = ch.ExchangeDeclare(
		"order_events", // name
		"topic",        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %v", err)
	}

	q, err := ch.QueueDeclare(
		"kitchen_order_queue", // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	err = ch.QueueBind(
		q.Name,          // queue name
		"order.created", // routing key
		"order_events",  // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind a queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	// 5. Build worker context avoiding rapid termination
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Println("Shutting down worker...")
		cancel()
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")

	for {
		select {
		case <-ctx.Done():
			return
		case d := <-msgs:
			log.Printf(" [x] Received %s", d.Body)
			var payload struct {
				OrderID uint   `json:"order_id"`
				Items   string `json:"items"`
			}
			if err := json.Unmarshal(d.Body, &payload); err == nil {
				// Translate event into new kitchen ticket
				ticket := &model.KitchenTicket{
					OrderID: payload.OrderID,
					Items:   payload.Items,
				}
				if err := kitchenSvc.CreateTicket(ticket); err != nil {
					log.Printf("[CorrID: %s] Failed to create ticket: %v", d.CorrelationId, err)
				} else {
					log.Printf("[CorrID: %s] Ticket created for order %d", d.CorrelationId, payload.OrderID)
				}
			} else {
				log.Printf("Error unmarshalling event string: %v", err)
			}
		}
	}
}
