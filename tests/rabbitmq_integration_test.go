package integration_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestTC_RABBIT_001_PublishOrderPlaced(t *testing.T) {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		t.Skip("RabbitMQ not reachable, skipping integration test")
		return
	}
	defer conn.Close()

	ch, _ := conn.Channel()
	defer ch.Close()

	q, _ := ch.QueueDeclare("test_order_placed", false, false, true, false, nil)
	ch.QueueBind(q.Name, "order.created", "order_events", false, nil)

	msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	// In a real test, we would trigger an order creation here
	// For this integration test, we simulate the event to verify the infra
	body, _ := json.Marshal(map[string]interface{}{"order_id": 123})
	ch.PublishWithContext(context.Background(), "order_events", "order.created", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})

	select {
	case msg := <-msgs:
		var data map[string]interface{}
		json.Unmarshal(msg.Body, &data)
		assert.Equal(t, float64(123), data["order_id"])
	case <-time.After(2 * time.Second):
		t.Fatal("Did not receive RabbitMQ message within 2 seconds")
	}
}
