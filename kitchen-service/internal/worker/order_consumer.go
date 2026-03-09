package worker

import (
	"encoding/json"
	"log"
	"os"

	"kitchen-service/internal/model"
	"kitchen-service/internal/service"

	amqp "github.com/rabbitmq/amqp091-go"
)

func StartOrderConsumer(kitchenSvc service.KitchenService) {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %v", err)
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Failed to open a channel: %v", err)
		return
	}

	// 1. ประกาศคิว
	q, err := ch.QueueDeclare("kitchen_order_queue", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to declare a queue: %v", err)
		return
	}

	// 2. Bind คิวเข้ากับ Exchange ตาม Key ที่สนใจ
	err = ch.QueueBind(q.Name, "order.created", "order_events", false, nil)
	if err != nil {
		log.Printf("Failed to bind queue: %v", err)
		return
	}

	// 3. เริ่มอ่านข้อความ
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to register a consumer: %v", err)
		return
	}

	go func() {
		log.Println("RabbitMQ Consumer started, waiting for orders...")
		for d := range msgs {
			var data map[string]interface{}
			if err := json.Unmarshal(d.Body, &data); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			// 4. นำข้อมูลไปสร้าง Ticket ใน DB ของ Kitchen
			orderIDFloat, ok := data["order_id"].(float64)
			if !ok {
				log.Printf("Invalid order_id missing or not float64: %v", data)
				continue
			}
			orderID := uint(orderIDFloat)

			itemsStr, _ := data["items"].(string)

			log.Printf("[CorrID: %s] Kitchen received new order: %v", d.CorrelationId, orderID)

			ticket := &model.KitchenTicket{
				OrderID: orderID,
				Items:   itemsStr,
			}

			if err := kitchenSvc.CreateTicket(ticket); err != nil {
				log.Printf("Failed to create ticket for order %d: %v", orderID, err)
			} else {
				log.Printf("Ticket successfully created for order %d", orderID)
			}
		}
	}()
}
