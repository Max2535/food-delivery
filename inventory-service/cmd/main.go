package main

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"inventory-service/internal/catalog"
	"inventory-service/internal/handler"
	"inventory-service/internal/middleware"
	"inventory-service/internal/model"
	"inventory-service/internal/repository"
	"inventory-service/internal/service"
	_ "inventory-service/docs" // Swagger docs
	swagger "github.com/gofiber/swagger"

	fiberprometheus "github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Inventory Service API
// @version 1.0
// @description This is the API server for an inventory service.
// @host localhost:3004
// @BasePath /

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Warn().Msg("Warning: .env file not found")
	}

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
		log.Fatal().Err(err).Msg("Could not connect to database after retries")
	}

	db.AutoMigrate(&model.RawMaterial{}, &model.StockTransaction{})

	// Catalog client for BOM lookup
	catalogURL := os.Getenv("CATALOG_SERVICE_URL")
	if catalogURL == "" {
		catalogURL = "http://catalog-service-api:3003"
	}
	catalogClient := catalog.NewClient(catalogURL)

	// Repositories
	materialRepo := repository.NewMaterialRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Services
	materialSvc := service.NewMaterialService(materialRepo)
	stockSvc := service.NewStockService(materialRepo, transactionRepo, catalogClient)

	// Handlers
	materialHandler := handler.NewMaterialHandler(materialSvc)
	stockHandler := handler.NewStockHandler(stockSvc)
	transactionHandler := handler.NewTransactionHandler(transactionRepo)

	app := fiber.New()
	app.Use(middleware.LoggerMiddleware())

	prome := fiberprometheus.NewWithDefaultRegistry("inventory-service")
	app.Use(prome.Middleware)
	prome.RegisterAt(app, "/metrics")

	// Swagger route
	app.Get("/swagger/*", swagger.HandlerDefault)

	api := app.Group("/api/v1/inventory")

	// Raw material routes
	materials := api.Group("/materials")
	materials.Get("/", materialHandler.GetAll)
	materials.Get("/low-stock", materialHandler.GetLowStock) // must be before /:id
	materials.Get("/:id", materialHandler.GetByID)
	materials.Post("/", materialHandler.Create)
	materials.Put("/:id", materialHandler.Update)

	// Stock operation routes
	stock := api.Group("/stock")
	stock.Post("/restock", stockHandler.Restock)
	stock.Post("/adjust", stockHandler.Adjust)
	stock.Post("/deduct", stockHandler.Deduct)

	// Transaction history routes
	transactions := api.Group("/transactions")
	transactions.Get("/", transactionHandler.GetAll)
	transactions.Get("/:material_id", transactionHandler.GetByMaterial)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3004"
	}

	log.Fatal().Err(app.Listen(":" + port)).Msg("Server failed")
}
