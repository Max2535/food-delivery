package main

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"catalog-service/internal/handler"
	"catalog-service/internal/middleware"
	"catalog-service/internal/model"
	"catalog-service/internal/repository"
	"catalog-service/internal/service"
	_ "catalog-service/docs" // Swagger docs
	swagger "github.com/gofiber/swagger"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Catalog Service API
// @version 1.0
// @description This is the API server for a catalog service.
// @host localhost:3003
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

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Warn().Err(err).Msg("Warning: Could not connect to Redis")
	}

	db.AutoMigrate(
		&model.MenuItem{},
		&model.Ingredient{},
		&model.BOMItem{},
		&model.KitchenStation{},
		&model.MenuStationMapping{},
		&model.BOMChoiceGroup{},
		&model.BOMChoiceOption{},
		&model.MenuAddOn{},
		&model.MenuPortionSize{},
	)

	// Seed ingredients
	seedIngredients := []model.Ingredient{
		{Name: "ข้าวสวย", Unit: "g"},
		{Name: "หมูสับ", Unit: "g"},
		{Name: "ใบกะเพรา", Unit: "g"},
		{Name: "กระเทียม", Unit: "g"},
		{Name: "พริกขี้หนู", Unit: "g"},
		{Name: "น้ำมันพืช", Unit: "ml"},
		{Name: "ไข่ไก่", Unit: "piece"},
		{Name: "น้ำปลา", Unit: "ml"},
		{Name: "น้ำตาลทราย", Unit: "g"},
		{Name: "ซอสหอยนางรม", Unit: "ml"},
		{Name: "เนื้อไก่", Unit: "g"},
		{Name: "ผักบุ้ง", Unit: "g"},
		{Name: "เส้นใหญ่", Unit: "g"},
		{Name: "กุ้งสด", Unit: "g"},
		{Name: "ปลากระพง", Unit: "g"},
		{Name: "ซีอิ๊วดำ", Unit: "ml"},
		{Name: "พริกแห้ง", Unit: "g"},
		{Name: "มะนาว", Unit: "piece"},
		{Name: "ผักกาดดอง", Unit: "g"},
		{Name: "ถั่วลิสง", Unit: "g"},
	}
	for i := range seedIngredients {
		var existing model.Ingredient
		if err := db.Where("name = ?", seedIngredients[i].Name).First(&existing).Error; err != nil {
			db.Create(&seedIngredients[i])
			log.Info().Str("ingredient", seedIngredients[i].Name).Msg("Seeded ingredient")
		}
	}

	// Seed menu items
	seedMenus := []model.MenuItem{
		{Name: "ข้าวผัดกระเพราหมูสับ", Price: 65, Category: "อาหารจานเดียว"},
		{Name: "ข้าวผัดกระเพราไก่", Price: 65, Category: "อาหารจานเดียว"},
		{Name: "ผัดไทกุ้งสด", Price: 85, Category: "อาหารจานเดียว"},
		{Name: "ผัดผักบุ้งไฟแดง", Price: 60, Category: "อาหารผัด"},
		{Name: "ปลากระพงทอดน้ำปลา", Price: 180, Category: "อาหารทอด"},
	}
	for i := range seedMenus {
		var existing model.MenuItem
		if err := db.Where("name = ?", seedMenus[i].Name).First(&existing).Error; err != nil {
			db.Create(&seedMenus[i])
			log.Info().Str("menu", seedMenus[i].Name).Msg("Seeded menu item")
		}
	}

	// Seed BOM (recipes) — link menu items to ingredients
	// Helper: find IDs after seeding
	ingredientIDByName := func(name string) *uint {
		var ing model.Ingredient
		if err := db.Where("name = ?", name).First(&ing).Error; err != nil {
			return nil
		}
		return &ing.ID
	}
	menuIDByName := func(name string) uint {
		var m model.MenuItem
		db.Where("name = ?", name).First(&m)
		return m.ID
	}

	type bomSeed struct {
		MenuName       string
		IngredientName string
		Quantity       float64
	}
	bomSeeds := []bomSeed{
		// ข้าวผัดกระเพราหมูสับ
		{"ข้าวผัดกระเพราหมูสับ", "ข้าวสวย", 250},
		{"ข้าวผัดกระเพราหมูสับ", "หมูสับ", 150},
		{"ข้าวผัดกระเพราหมูสับ", "ใบกะเพรา", 20},
		{"ข้าวผัดกระเพราหมูสับ", "กระเทียม", 15},
		{"ข้าวผัดกระเพราหมูสับ", "พริกขี้หนู", 10},
		{"ข้าวผัดกระเพราหมูสับ", "น้ำมันพืช", 30},
		{"ข้าวผัดกระเพราหมูสับ", "น้ำปลา", 15},
		{"ข้าวผัดกระเพราหมูสับ", "น้ำตาลทราย", 5},
		{"ข้าวผัดกระเพราหมูสับ", "ซอสหอยนางรม", 10},
		// ข้าวผัดกระเพราไก่
		{"ข้าวผัดกระเพราไก่", "ข้าวสวย", 250},
		{"ข้าวผัดกระเพราไก่", "เนื้อไก่", 150},
		{"ข้าวผัดกระเพราไก่", "ใบกะเพรา", 20},
		{"ข้าวผัดกระเพราไก่", "กระเทียม", 15},
		{"ข้าวผัดกระเพราไก่", "พริกขี้หนู", 10},
		{"ข้าวผัดกระเพราไก่", "น้ำมันพืช", 30},
		{"ข้าวผัดกระเพราไก่", "น้ำปลา", 15},
		{"ข้าวผัดกระเพราไก่", "น้ำตาลทราย", 5},
		{"ข้าวผัดกระเพราไก่", "ซอสหอยนางรม", 10},
		// ผัดไทกุ้งสด
		{"ผัดไทกุ้งสด", "เส้นใหญ่", 200},
		{"ผัดไทกุ้งสด", "กุ้งสด", 100},
		{"ผัดไทกุ้งสด", "ไข่ไก่", 1},
		{"ผัดไทกุ้งสด", "น้ำปลา", 20},
		{"ผัดไทกุ้งสด", "น้ำตาลทราย", 30},
		{"ผัดไทกุ้งสด", "ถั่วลิสง", 20},
		{"ผัดไทกุ้งสด", "ผักบุ้ง", 30},
		{"ผัดไทกุ้งสด", "มะนาว", 1},
		// ผัดผักบุ้งไฟแดง
		{"ผัดผักบุ้งไฟแดง", "ผักบุ้ง", 200},
		{"ผัดผักบุ้งไฟแดง", "กระเทียม", 20},
		{"ผัดผักบุ้งไฟแดง", "พริกแห้ง", 10},
		{"ผัดผักบุ้งไฟแดง", "น้ำมันพืช", 40},
		{"ผัดผักบุ้งไฟแดง", "ซอสหอยนางรม", 15},
		{"ผัดผักบุ้งไฟแดง", "น้ำปลา", 10},
		// ปลากระพงทอดน้ำปลา
		{"ปลากระพงทอดน้ำปลา", "ปลากระพง", 400},
		{"ปลากระพงทอดน้ำปลา", "กระเทียม", 30},
		{"ปลากระพงทอดน้ำปลา", "พริกขี้หนู", 15},
		{"ปลากระพงทอดน้ำปลา", "น้ำปลา", 30},
		{"ปลากระพงทอดน้ำปลา", "น้ำมันพืช", 500},
		{"ปลากระพงทอดน้ำปลา", "มะนาว", 2},
	}
	for _, b := range bomSeeds {
		menuID := menuIDByName(b.MenuName)
		ingID := ingredientIDByName(b.IngredientName)
		if menuID == 0 || ingID == nil {
			continue
		}
		var existing model.BOMItem
		if err := db.Where("menu_item_id = ? AND ingredient_id = ?", menuID, *ingID).First(&existing).Error; err != nil {
			db.Create(&model.BOMItem{
				MenuItemID:   menuID,
				IngredientID: ingID,
				Quantity:     b.Quantity,
			})
			log.Info().Str("menu", b.MenuName).Str("ingredient", b.IngredientName).Msg("Seeded BOM item")
		}
	}

	// Repositories
	menuRepo := repository.NewMenuRepository(db)
	ingredientRepo := repository.NewIngredientRepository(db)
	bomRepo := repository.NewBOMRepository(db)
	stationRepo := repository.NewStationRepository(db)
	choiceRepo := repository.NewChoiceRepository(db)
	addOnRepo := repository.NewAddOnRepository(db)
	portionRepo := repository.NewPortionRepository(db)

	// Services
	menuSvc := service.NewMenuService(menuRepo, redisClient)
	ingredientSvc := service.NewIngredientService(ingredientRepo)
	bomSvc := service.NewBOMService(bomRepo, ingredientRepo, menuRepo)
	stationSvc := service.NewStationService(stationRepo, menuRepo)
	choiceSvc := service.NewChoiceService(choiceRepo, ingredientRepo)
	addOnSvc := service.NewAddOnService(addOnRepo, ingredientRepo)
	portionSvc := service.NewPortionService(portionRepo, menuRepo)

	// Handlers
	menuHandler := handler.NewMenuHandler(menuSvc, stationSvc)
	ingredientHandler := handler.NewIngredientHandler(ingredientSvc)
	bomHandler := handler.NewBOMHandler(bomSvc)
	stationHandler := handler.NewStationHandler(stationSvc)
	choiceHandler := handler.NewChoiceHandler(choiceSvc)
	addOnHandler := handler.NewAddOnHandler(addOnSvc)
	portionHandler := handler.NewPortionHandler(portionSvc)

	app := fiber.New()

	app.Use(middleware.LoggerMiddleware())

	prome := fiberprometheus.NewWithDefaultRegistry("catalog-service")
	app.Use(prome.Middleware)
	prome.RegisterAt(app, "/metrics")

	// Swagger route
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Routes
	api := app.Group("/api/v1/catalog")

	// Menu endpoints
	menus := api.Group("/menus")
	menus.Get("/", menuHandler.GetAllMenuItems)
	menus.Get("/:id", menuHandler.GetMenuItemByID)
	menus.Post("/", menuHandler.CreateMenuItem)
	menus.Put("/:id", menuHandler.UpdateMenuItem)
	menus.Delete("/:id", menuHandler.DeleteMenuItem)

	// BOM endpoints (nested under menus)
	menus.Get("/:id/bom", bomHandler.GetBOM)
	menus.Get("/:id/bom/flat", bomHandler.GetFlatBOM)
	menus.Post("/:id/bom", bomHandler.AddBOMItem)
	menus.Delete("/:id/bom/:bom_id", bomHandler.DeleteBOMItem)

	// Choice Group endpoints (Case 1: customer selects ingredient)
	menus.Get("/:id/choices", choiceHandler.GetChoices)
	menus.Post("/:id/choices", choiceHandler.CreateChoiceGroup)
	menus.Delete("/:id/choices/:group_id", choiceHandler.DeleteChoiceGroup)
	menus.Post("/:id/choices/:group_id/options", choiceHandler.AddChoiceOption)
	menus.Delete("/:id/choices/:group_id/options/:option_id", choiceHandler.DeleteChoiceOption)

	// Add-on endpoints (Case 2: optional extras)
	menus.Get("/:id/addons", addOnHandler.GetAddOns)
	menus.Post("/:id/addons", addOnHandler.CreateAddOn)
	menus.Delete("/:id/addons/:addon_id", addOnHandler.DeleteAddOn)

	// Portion size endpoints (Case 3: size variants)
	menus.Get("/:id/portions", portionHandler.GetPortions)
	menus.Post("/:id/portions", portionHandler.CreatePortion)
	menus.Delete("/:id/portions/:portion_id", portionHandler.DeletePortion)

	// Menu → Station assignment
	menus.Post("/:id/station", stationHandler.AssignMenuToStation)

	// Ingredient endpoints
	ingredients := api.Group("/ingredients")
	ingredients.Get("/", ingredientHandler.GetAllIngredients)
	ingredients.Post("/", ingredientHandler.CreateIngredient)

	// Station endpoints
	stations := api.Group("/stations")
	stations.Get("/", stationHandler.GetAllStations)
	stations.Post("/", stationHandler.CreateStation)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3003"
	}

	log.Fatal().Err(app.Listen(":" + port)).Msg("Server failed")
}
