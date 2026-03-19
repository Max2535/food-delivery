package main

import (
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"auth-service/internal/model"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"os"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	_ "auth-service/docs" // Swagger docs
	swagger "github.com/gofiber/swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Auth Service API
// @version 1.0
// @description This is the API server for an auth service.
// @host localhost:3005
// @BasePath /

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg(".env file not found, using environment variables")
	}

	// Database Connection
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal().Msg("DB_URL is not set")
	}

	var db *gorm.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
		if err == nil {
			break
		}
		log.Warn().Err(err).Int("attempt", i+1).Msg("Failed to connect to database")
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to database")
	}

	// Auto Migrate
	db.AutoMigrate(&model.User{})

	// Initialize Layers
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authSvc)

	// Seed test users for TestSprite tests
	log.Info().Msg("Seeding validUser...")
	if err := authSvc.Register(&model.User{Username: "validUser", Password: "validPassword123!", Email: "validuser@example.com"}); err != nil {
		log.Fatal().Err(err).Msg("Failed to seed validUser")
	}
	log.Info().Msg("Seeding testuser...")
	if err := authSvc.Register(&model.User{Username: "testuser", Password: "TestPass123!", Email: "testuser@example.com"}); err != nil {
		log.Fatal().Err(err).Msg("Failed to seed testuser")
	}
	log.Info().Msg("Seeding seeded_user...")
	if err := authSvc.Register(&model.User{Username: "seeded_user", Password: "seeded_password", Email: "seeded_user@example.com"}); err != nil {
		log.Fatal().Err(err).Msg("Failed to seed seeded_user")
	}

	// Fiber Instance
	app := fiber.New()

	// Global Middleware
	app.Use(middleware.LoggerMiddleware())

	// 1. Prometheus Middleware - ใช้เป้าหมายเดียวกับ order-service
	prometheus := fiberprometheus.NewWithDefaultRegistry("auth-service")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	// Swagger route
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Routes
	api := app.Group("/api")
	v1 := api.Group("/v1")

	auth := v1.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/register", authHandler.Register)

	// Listen
	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}

	log.Info().Str("port", port).Msg("Auth Service starting...")
	log.Fatal().Err(app.Listen(":" + port)).Msg("Server failed")
}
