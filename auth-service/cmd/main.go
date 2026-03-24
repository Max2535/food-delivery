package main

import (
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"auth-service/internal/model"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/internal/telemetry"
	"context"

	_ "auth-service/docs"
	"os"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	swagger "github.com/gofiber/swagger"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Auth Service API
// @version 1.0
// @description JWT-based authentication service
// @host localhost:3005
// @BasePath /
func main() {
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg(".env file not found, using environment variables")
	}

	// OpenTelemetry
	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint != "" {
		shutdown := telemetry.InitTracer("auth-service", otelEndpoint)
		defer shutdown(context.Background())
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal().Msg("DB_URL is not set")
	}

	var db *gorm.DB
	var err error
	for i := range 10 {
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

	if err := db.AutoMigrate(&model.Role{}, &model.Group{}, &model.User{}, &model.RefreshToken{}, &model.PasswordResetToken{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to auto migrate")
	}

	// Seed Roles
	roleMap := make(map[string]model.Role)
	for _, roleName := range []string{model.RoleAdmin, model.RoleRider, model.RoleCustomer, model.RoleUser, model.RoleMerchant} {
		var role model.Role
		if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				role = model.Role{Name: roleName}
				db.Create(&role)
				log.Info().Str("role", roleName).Msg("Seeded role successfully")
			}
		}
		roleMap[roleName] = role
	}

	// Seed Groups (each group maps to a set of roles)
	groupRoles := map[string][]string{
		model.GroupUser:     {model.RoleUser},
		model.GroupCustomer: {model.RoleCustomer, model.RoleUser},
		model.GroupRider:    {model.RoleRider, model.RoleUser},
		model.GroupMerchant: {model.RoleMerchant, model.RoleUser},
		model.GroupAdmin:    {model.RoleAdmin, model.RoleMerchant, model.RoleRider, model.RoleCustomer, model.RoleUser},
	}
	for groupName, roleNames := range groupRoles {
		var group model.Group
		if err := db.Where("name = ?", groupName).First(&group).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				roles := make([]model.Role, len(roleNames))
				for i, rn := range roleNames {
					roles[i] = roleMap[rn]
				}
				group = model.Group{Name: groupName, Roles: roles}
				db.Create(&group)
				log.Info().Str("group", groupName).Int("roles", len(roles)).Msg("Seeded group successfully")
			}
		}
	}

	// Initialize Layers
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewRefreshTokenRepository(db)
	resetTokenRepo := repository.NewPasswordResetTokenRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	authSvc := service.NewAuthService(userRepo, tokenRepo, resetTokenRepo, groupRepo)
	authHandler := handler.NewAuthHandler(authSvc)

	// Seed test users
	testUsers := []struct{ username, password, email string }{
		{"admin", "adminpassword", "admin@food-delivery.com"},
		{"rider_01", "securepassword123", "rider01@food-delivery.com"},
		{"customer_01", "password123", "customer@food-delivery.com"},
		{"validuser", "validpassword", "validuser@example.com"},
	}
	for _, u := range testUsers {
		if _, err := authSvc.Register(u.username, u.password, u.email); err != nil {
			log.Warn().Err(err).Str("username", u.username).Msg("Could not seed user (may already exist)")
		} else {
			log.Info().Str("username", u.username).Msg("Seeded user successfully")
		}
	}

	// Fiber Instance
	app := fiber.New()

	// Global Middleware
	app.Use(otelfiber.Middleware())
	app.Use(middleware.LoggerMiddleware())

	// 1. Prometheus Middleware - ใช้เป้าหมายเดียวกับ order-service
	prometheus := fiberprometheus.NewWithDefaultRegistry("auth-service")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	// Swagger route
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Routes
	auth := app.Group("/api/v1/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)
	auth.Post("/logout-all", authHandler.LogoutAll)
	auth.Get("/profile", authHandler.GetProfile)
	auth.Put("/password", authHandler.ChangePassword)
	auth.Post("/forgot-password", authHandler.ForgotPassword)
	auth.Post("/reset-password", authHandler.ResetPassword)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}
	log.Info().Str("port", port).Msg("Auth Service starting...")
	log.Fatal().Err(app.Listen(":" + port)).Msg("Server failed")
}
