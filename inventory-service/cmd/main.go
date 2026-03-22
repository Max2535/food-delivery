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

	// Seed raw materials (CatalogIngredientID must match catalog-service Ingredient IDs)
	type materialSeed struct {
		CatalogIngredientID uint
		Name                string
		Unit                string
		CurrentStock        float64
		ReorderPoint        float64
	}
	seeds := []materialSeed{
		{1, "ข้าวสวย", "g", 50000, 5000},
		{2, "หมูสับ", "g", 20000, 3000},
		{3, "ใบกะเพรา", "g", 5000, 500},
		{4, "กระเทียม", "g", 10000, 1000},
		{5, "พริกขี้หนู", "g", 3000, 500},
		{6, "น้ำมันพืช", "ml", 20000, 3000},
		{7, "ไข่ไก่", "piece", 500, 50},
		{8, "น้ำปลา", "ml", 10000, 1000},
		{9, "น้ำตาลทราย", "g", 10000, 1000},
		{10, "ซอสหอยนางรม", "ml", 5000, 500},
		{11, "เนื้อไก่", "g", 20000, 3000},
		{12, "ผักบุ้ง", "g", 10000, 1000},
		{13, "เส้นใหญ่", "g", 15000, 2000},
		{14, "กุ้งสด", "g", 10000, 1500},
		{15, "ปลากระพง", "g", 8000, 1000},
		{16, "ซีอิ๊วดำ", "ml", 5000, 500},
		{17, "พริกแห้ง", "g", 3000, 300},
		{18, "มะนาว", "piece", 200, 30},
		{19, "ผักกาดดอง", "g", 5000, 500},
		{20, "ถั่วลิสง", "g", 5000, 500},
	}
	for _, s := range seeds {
		var existing model.RawMaterial
		if err := db.Where("name = ?", s.Name).First(&existing).Error; err != nil {
			catID := s.CatalogIngredientID
			db.Create(&model.RawMaterial{
				CatalogIngredientID: &catID,
				Name:                s.Name,
				Unit:                s.Unit,
				CurrentStock:        s.CurrentStock,
				ReorderPoint:        s.ReorderPoint,
			})
			log.Info().Str("material", s.Name).Msg("Seeded raw material")
		}
	}

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
