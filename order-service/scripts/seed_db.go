package main

import (
	"log"
	"os"

	"order-service/internal/model"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Database Connection
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not set in .env")
	}

	log.Printf("Connecting to database at %s...", dbURL)
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection successful.")

	// AutoMigrate the schema
	log.Println("Migrating database...")
	if err := db.AutoMigrate(&model.Order{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Mock Data
	orders := []model.Order{
		{
			CustomerID:  "CUST001",
			TotalAmount: 250.50,
			Status:      "pending",
		},
		{
			CustomerID:  "CUST002",
			TotalAmount: 120.00,
			Status:      "completed",
		},
		{
			CustomerID:  "CUST003",
			TotalAmount: 50.00,
			Status:      "cancelled",
		},
		{
			CustomerID:  "CUST001",
			TotalAmount: 320.75,
			Status:      "completed",
		},
		{
			CustomerID:  "CUST004",
			TotalAmount: 450.00,
			Status:      "preparing",
		},
	}

	log.Println("Seeding data...")

	for _, order := range orders {
		if err := db.Create(&order).Error; err != nil {
			log.Printf("Failed to seed order for %s: %v", order.CustomerID, err)
		} else {
			log.Printf("Seeded order ID %d for customer %s", order.ID, order.CustomerID)
		}
	}

	log.Println("Seeding complete!")
}
