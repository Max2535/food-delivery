package main

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"

	_ "order-service/docs" // Swagger docs
	"order-service/internal/handler"
	"order-service/internal/middleware"
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
		log.Warn().Msg("Warning: .env file not found")
	}

	// Database Connection
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal().Msg("DB_URL is not set in .env")
	}

	var db *gorm.DB
	var err error

	// Retry connection up to 10 times
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

	// Auto Migrate
	db.AutoMigrate(&model.Order{}, &model.OrderItem{})

	// Initialize Layers
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	// Fiber Instance
	app := fiber.New()

	// Global Middleware
	app.Use(middleware.LoggerMiddleware())

	// 1. ใช้ Middleware แบบ NewWithDefaultRegistry เพื่อรวบรวม Go Runtime Metrics อัตโนมัติ
	prome := fiberprometheus.NewWithDefaultRegistry("order-service")

	// 2. ลงทะเบียน Middleware เข้ากับ Fiber
	app.Use(prome.Middleware)

	// 3.Newman Tip: แทนที่จะให้มันสร้าง Registry เอง
	// เราจะใช้ Default ของ Prometheus ที่รวม Go Metrics ไว้แล้ว
	prome.RegisterAt(app, "/metrics")

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

	log.Fatal().Err(app.Listen(":" + port)).Msg("Server failed")
}
