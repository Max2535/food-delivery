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

	if err := db.AutoMigrate(&model.Permission{}, &model.Role{}, &model.Group{}, &model.User{}, &model.RefreshToken{}, &model.PasswordResetToken{}, &model.NavGroup{}, &model.NavItem{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to auto migrate")
	}

	// Seed Permissions
	permMap := make(map[string]model.Permission)
	for _, p := range model.AllPermissions {
		var perm model.Permission
		if err := db.Where("name = ?", p.Name).First(&perm).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				perm = model.Permission{Name: p.Name, Description: p.Description}
				db.Create(&perm)
				log.Info().Str("permission", p.Name).Msg("Seeded permission")
			}
		}
		permMap[p.Name] = perm
	}

	// Seed Roles
	roleMap := make(map[string]model.Role)
	for _, roleName := range []string{model.RoleAdmin, model.RoleRider, model.RoleCustomer, model.RoleUser, model.RoleMerchant} {
		var role model.Role
		if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				role = model.Role{Name: roleName}
				db.Create(&role)
				log.Info().Str("role", roleName).Msg("Seeded role")
			}
		}
		// Assign permissions to role (idempotent — replace)
		if permNames, ok := model.RolePermissions[roleName]; ok {
			perms := make([]model.Permission, 0, len(permNames))
			for _, pn := range permNames {
				if p, ok := permMap[pn]; ok {
					perms = append(perms, p)
				}
			}
			db.Model(&role).Association("Permissions").Replace(perms)
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
				log.Info().Str("group", groupName).Int("roles", len(roles)).Msg("Seeded group")
			}
		}
	}

	// Seed Nav Menu Config
	for i, seed := range model.DefaultNavMenuSeed {
		var existing model.NavGroup
		if err := db.Where("label = ?", seed.Label).First(&existing).Error; err == nil {
			continue // already seeded
		}
		groupPerms := make([]model.Permission, 0, len(seed.Permissions))
		for _, pn := range seed.Permissions {
			if p, ok := permMap[pn]; ok {
				groupPerms = append(groupPerms, p)
			}
		}
		navGroup := model.NavGroup{Label: seed.Label, SortOrder: i, Permissions: groupPerms}
		if err := db.Create(&navGroup).Error; err != nil {
			log.Error().Err(err).Str("label", seed.Label).Msg("Failed to seed nav group")
			continue
		}
		for j, si := range seed.Items {
			itemPerms := make([]model.Permission, 0, len(si.Permissions))
			for _, pn := range si.Permissions {
				if p, ok := permMap[pn]; ok {
					itemPerms = append(itemPerms, p)
				}
			}
			navItem := model.NavItem{NavGroupID: navGroup.ID, Label: si.Label, Href: si.Href, SortOrder: j, Permissions: itemPerms}
			if err := db.Create(&navItem).Error; err != nil {
				log.Error().Err(err).Str("label", si.Label).Msg("Failed to seed nav item")
			}
		}
		log.Info().Str("label", seed.Label).Int("items", len(seed.Items)).Msg("Seeded nav group")
	}

	// Initialize Layers
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewRefreshTokenRepository(db)
	resetTokenRepo := repository.NewPasswordResetTokenRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	navMenuRepo := repository.NewNavMenuRepository(db)
	authSvc := service.NewAuthService(userRepo, tokenRepo, resetTokenRepo, groupRepo, navMenuRepo)
	authHandler := handler.NewAuthHandler(authSvc)

	// Seed test users
	testUsers := []struct{ username, password, email string }{
		{"admin", "admin", "admin@food-delivery.com"},
		// {"rider_01", "securepassword123", "rider01@food-delivery.com"},
		// {"customer_01", "password123", "customer@food-delivery.com"},
		// {"validuser", "validpassword", "validuser@example.com"},
	}
	for _, u := range testUsers {
		user, err := authSvc.Register(u.username, u.password, u.email)
		if err != nil {
			log.Warn().Err(err).Str("username", u.username).Msg("Could not seed user (may already exist)")
			continue
		}
		// Assign admin group
		adminGroup := model.Group{}
		if err := db.Where("name = ?", model.GroupAdmin).Preload("Roles").First(&adminGroup).Error; err == nil {
			user.GroupID = adminGroup.ID
			user.Group = adminGroup
			db.Save(user)
			log.Info().Str("username", u.username).Str("group", model.GroupAdmin).Msg("Seeded user with admin group")
		} else {
			log.Warn().Err(err).Str("username", u.username).Msg("Admin group not found, user keeps default group")
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
	auth.Get("/groups", authHandler.ListGroups)
	auth.Post("/group", authHandler.CreateGroup)
	auth.Put("/group", authHandler.UpdateGroup)
	auth.Delete("/group/:id", authHandler.DeleteGroup)
	auth.Get("/roles", authHandler.ListRoles)
	auth.Get("/menu-config", authHandler.GetMenuConfig)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}
	log.Info().Str("port", port).Msg("Auth Service starting...")
	log.Fatal().Err(app.Listen(":" + port)).Msg("Server failed")
}
