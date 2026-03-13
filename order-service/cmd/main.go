package main

import (
	"log"
	"os"
	"time"

	_ "order-service/docs" // Swagger docs
	"order-service/internal/handler"
	"order-service/internal/model"
	"order-service/internal/repository"
	"order-service/internal/service"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	swagger "github.com/gofiber/swagger"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Order Service API
// @version 1.0
// @description This is the API server for an order service.
// @host localhost:3000
// @BasePath /

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

	var db *gorm.DB
	var err error

	// Retry connection up to 10 times
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database (attempt %d/10): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to database after retries: %v", err)
	}

	// Auto Migrate
	db.AutoMigrate(&model.Order{})

	// Initialize Layers
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	// Fiber Instance
	app := fiber.New()

	// Setup Prometheus
	prometheus := fiberprometheus.New("order-service")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	// Swagger route
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Routes
	api := app.Group("/api")
	v1 := api.Group("/v1")

	orders := v1.Group("/orders")
	orders.Post("/", orderHandler.CreateOrder)
	orders.Get("/", orderHandler.GetAllOrders)
	orders.Get("/:id", orderHandler.GetOrderByID)

	// Listen
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}
