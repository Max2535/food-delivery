package main

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"catalog-service/internal/handler"
	"catalog-service/internal/middleware"
	"catalog-service/internal/model"
	"catalog-service/internal/repository"
	"catalog-service/internal/service"

	fiberprometheus "github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

	db.AutoMigrate(
		&model.MenuItem{},
		&model.Ingredient{},
		&model.BOMItem{},
		&model.KitchenStation{},
		&model.MenuStationMapping{},
	)

	// Repositories
	menuRepo := repository.NewMenuRepository(db)
	ingredientRepo := repository.NewIngredientRepository(db)
	bomRepo := repository.NewBOMRepository(db)
	stationRepo := repository.NewStationRepository(db)

	// Services
	menuSvc := service.NewMenuService(menuRepo)
	ingredientSvc := service.NewIngredientService(ingredientRepo)
	bomSvc := service.NewBOMService(bomRepo, ingredientRepo)
	stationSvc := service.NewStationService(stationRepo, menuRepo)

	// Handlers
	menuHandler := handler.NewMenuHandler(menuSvc, stationSvc)
	ingredientHandler := handler.NewIngredientHandler(ingredientSvc)
	bomHandler := handler.NewBOMHandler(bomSvc)
	stationHandler := handler.NewStationHandler(stationSvc)

	app := fiber.New()

	app.Use(middleware.LoggerMiddleware())

	prome := fiberprometheus.NewWithDefaultRegistry("catalog-service")
	app.Use(prome.Middleware)
	prome.RegisterAt(app, "/metrics")

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
	menus.Post("/:id/bom", bomHandler.AddBOMItem)
	menus.Delete("/:id/bom/:bom_id", bomHandler.DeleteBOMItem)

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
