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
	if err := db.AutoMigrate(&model.Order{}, &model.OrderItem{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Mock Data with Items
	orders := []model.Order{
		{
			CustomerID:  "CUST001",
			TotalAmount: 220.00,
			Status:      "pending",
			Items: []model.OrderItem{
				{MenuItemID: 1, MenuItemName: "ผัดกะเพราหมูสับ", UnitPrice: 50.00, Quantity: 2, TotalPrice: 100.00},
				{MenuItemID: 2, MenuItemName: "ต้มยำกุ้ง", UnitPrice: 120.00, Quantity: 1, TotalPrice: 120.00},
			},
		},
		{
			CustomerID:  "CUST002",
			TotalAmount: 170.00,
			Status:      "completed",
			Items: []model.OrderItem{
				{MenuItemID: 3, MenuItemName: "ข้าวมันไก่", UnitPrice: 50.00, Quantity: 1, TotalPrice: 50.00},
				{MenuItemID: 2, MenuItemName: "ต้มยำกุ้ง", UnitPrice: 120.00, Quantity: 1, TotalPrice: 120.00},
			},
		},
		{
			CustomerID:  "CUST003",
			TotalAmount: 50.00,
			Status:      "cancelled",
			Items: []model.OrderItem{
				{MenuItemID: 1, MenuItemName: "ผัดกะเพราหมูสับ", UnitPrice: 50.00, Quantity: 1, TotalPrice: 50.00},
			},
		},
		{
			CustomerID:  "CUST001",
			TotalAmount: 320.75,
			Status:      "completed",
			Items: []model.OrderItem{
				{MenuItemID: 4, MenuItemName: "แกงเขียวหวาน", UnitPrice: 80.25, Quantity: 1, TotalPrice: 80.25},
				{MenuItemID: 5, MenuItemName: "ส้มตำ", UnitPrice: 60.00, Quantity: 2, TotalPrice: 120.00},
				{MenuItemID: 2, MenuItemName: "ต้มยำกุ้ง", UnitPrice: 120.50, Quantity: 1, TotalPrice: 120.50},
			},
		},
		{
			CustomerID:  "CUST004",
			TotalAmount: 450.00,
			Status:      "preparing",
			Items: []model.OrderItem{
				{MenuItemID: 6, MenuItemName: "ปลากะพงทอดน้ำปลา", UnitPrice: 250.00, Quantity: 1, TotalPrice: 250.00},
				{MenuItemID: 7, MenuItemName: "ผัดผักรวมมิตร", UnitPrice: 100.00, Quantity: 2, TotalPrice: 200.00},
			},
		},
	}

	log.Println("Seeding data...")

	for _, order := range orders {
		if err := db.Create(&order).Error; err != nil {
			log.Printf("Failed to seed order for %s: %v", order.CustomerID, err)
		} else {
			log.Printf("Seeded order ID %d for customer %s (%d items)", order.ID, order.CustomerID, len(order.Items))
		}
	}

	log.Println("Seeding complete!")
}
