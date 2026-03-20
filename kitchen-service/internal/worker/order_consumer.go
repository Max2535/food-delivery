package worker

import (
	"encoding/json"
	"os"

	"github.com/rs/zerolog/log"

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
		log.Error().Err(err).Msg("Failed to connect to RabbitMQ")
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Error().Err(err).Msg("Failed to open a channel")
		return
	}

	// 1. ประกาศคิว
	q, err := ch.QueueDeclare("kitchen_order_queue", true, false, false, false, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to declare a queue")
		return
	}

	// 2. Bind คิวเข้ากับ Exchange ตาม Key ที่สนใจ
	err = ch.QueueBind(q.Name, "order.created", "order_events", false, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to bind queue")
		return
	}

	// 3. เริ่มอ่านข้อความ
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to register a consumer")
		return
	}

	go func() {
		log.Info().Str("service", "kitchen-service").Msg("RabbitMQ Consumer started, waiting for orders...")
		for d := range msgs {
			var data map[string]interface{}
			if err := json.Unmarshal(d.Body, &data); err != nil {
				log.Error().Err(err).Msg("Error unmarshaling message")
				continue
			}

			// 4. นำข้อมูลไปสร้าง Ticket ใน DB ของ Kitchen
			var msgData struct {
				OrderID uint            `json:"order_id"`
				Items   json.RawMessage `json:"items"`
			}
			if err := json.Unmarshal(d.Body, &msgData); err != nil {
				log.Error().Err(err).Msg("Error unmarshaling message body")
				continue
			}

			correlationID := d.CorrelationId
			if correlationID == "" {
				correlationID = "unknown"
			}

			log.Info().
				Str("service", "kitchen-service").
				Str("correlation_id", correlationID).
				Uint("order_id", msgData.OrderID).
				Msg("Kitchen received new order")

			ticket := &model.KitchenTicket{
				OrderID: msgData.OrderID,
				Items:   string(msgData.Items),
			}

			if err := kitchenSvc.CreateTicket(ticket); err != nil {
				log.Error().Err(err).Uint("order_id", msgData.OrderID).Str("correlation_id", correlationID).Msg("Failed to create ticket for order")
			} else {
				log.Info().Uint("order_id", msgData.OrderID).Str("correlation_id", correlationID).Msg("Ticket successfully created for order")
			}
		}
	}()
}
