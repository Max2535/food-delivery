package main

import (
	"log"
	"os"

	"kitchen-service/internal/model"

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
	if err := db.AutoMigrate(&model.KitchenTicket{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Mock Data
	tickets := []model.KitchenTicket{
		{
			OrderID: 1,
			Items:   `[{"name": "Pad Thai", "quantity": 1}, {"name": "Tom Yum", "quantity": 1}]`,
			Status:  "Received",
		},
		{
			OrderID: 2,
			Items:   `[{"name": "Green Curry", "quantity": 2}]`,
			Status:  "Cooking",
		},
		{
			OrderID: 3,
			Items:   `[{"name": "Mango Sticky Rice", "quantity": 1}]`,
			Status:  "Ready",
		},
	}

	log.Println("Seeding data...")

	for _, ticket := range tickets {
		if err := db.Create(&ticket).Error; err != nil {
			log.Printf("Failed to seed ticket for order %d: %v", ticket.OrderID, err)
		} else {
			log.Printf("Seeded ticket ID %d for order %d", ticket.ID, ticket.OrderID)
		}
	}

	log.Println("Seeding complete!")
}
