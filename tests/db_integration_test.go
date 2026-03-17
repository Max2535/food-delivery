package integration_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Define a local struct for testing persistence to avoid complex module imports
type TestOrder struct {
	ID              uint      `gorm:"primaryKey;autoIncrement"`
	CustomerID      string    `gorm:"type:varchar(100);not null"`
	TotalAmount     float64   `gorm:"type:decimal(10,2)"`
	Status          string    `gorm:"type:varchar(20);default:'Pending'"`
	DeliveryAddress string    `gorm:"type:text"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (TestOrder) TableName() string {
	return "orders"
}

func TestTC_DB_001_OrderPersistence(t *testing.T) {
	// Connection string for local host to Docker Postgres
	dsn := "host=localhost user=admin password=admin dbname=order_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Actual Database not reachable: %v. Skipping integration test.", err)
		return
	}

	// Ensure the table exists (in case AutoMigrate hasn't run yet in the service)
	db.AutoMigrate(&TestOrder{})

	order := TestOrder{
		CustomerID:      "TEST_PERSIST_001",
		TotalAmount:     250.75,
		Status:          "Pending",
		DeliveryAddress: "123 Test Street, Integration City",
	}

	// 1. Create - INSERT into DB
	err = db.Create(&order).Error
	assert.NoError(t, err)
	assert.NotZero(t, order.ID)

	// 2. Read - FETCH from DB
	var fetchedOrder TestOrder
	err = db.First(&fetchedOrder, order.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, order.CustomerID, fetchedOrder.CustomerID)
	assert.Equal(t, order.DeliveryAddress, fetchedOrder.DeliveryAddress)

	t.Logf("--- Fetched Data from DB ---")
	t.Logf("Order ID: %d", fetchedOrder.ID)
	t.Logf("Customer: %s", fetchedOrder.CustomerID)
	t.Logf("Amount:   %.2f", fetchedOrder.TotalAmount)
	t.Logf("Status:   %s", fetchedOrder.Status)
	t.Logf("Address:  %s", fetchedOrder.DeliveryAddress)
	t.Logf("Created:  %s", fetchedOrder.CreatedAt.Format(time.RFC3339))
	t.Logf("----------------------------")

	t.Logf("Successfully inserted and verified order ID: %d in order_db", order.ID)
}
