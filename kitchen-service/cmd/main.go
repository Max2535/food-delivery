package main

import (
	"context"
	"kitchen-service/internal/handler"
	"kitchen-service/internal/middleware"
	"kitchen-service/internal/model"
	"kitchen-service/internal/repository"
	"kitchen-service/internal/service"
	"kitchen-service/internal/telemetry"
	"kitchen-service/internal/worker"
	_ "kitchen-service/docs" // Swagger docs
	"os"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	swagger "github.com/gofiber/swagger"
	consul "github.com/hashicorp/consul/api"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Kitchen Service API
// @version 1.0
// @description This is the API server for a kitchen service.
// @host localhost:3005
// @BasePath /

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Warn().Msg("Warning: .env file not found")
	}

	// OpenTelemetry
	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint != "" {
		shutdown := telemetry.InitTracer("kitchen-service", otelEndpoint)
		defer shutdown(context.Background())
	}

	// 1. Database Connection (Kitchen DB)
    dsn := os.Getenv("DB_URL") 
    db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    
    // Auto Migrate
    db.AutoMigrate(&model.KitchenTicket{})
    
    // 2. Setup Layers
    repo := repository.NewKitchenRepository(db)
    svc := service.NewKitchenService(repo)

    // เริ่มทำงาน Consumer ให้รอรับ RabbitMQ Events 
    worker.StartOrderConsumer(svc)

    hdl := handler.NewKitchenHandler(svc)

    app := fiber.New()

	// Global Middleware
	app.Use(otelfiber.Middleware())
	app.Use(middleware.LoggerMiddleware())

    // 1. ใช้ Middleware แบบ NewWithDefaultRegistry เพื่อรวบรวม Go Runtime Metrics อัตโนมัติ
    prome := fiberprometheus.NewWithDefaultRegistry("kitchen-service")

    // 2. ลงทะเบียน Middleware เข้ากับ Fiber
    app.Use(prome.Middleware)

    // 3.Newman Tip: แทนที่จะให้มันสร้าง Registry เอง
    // เราจะใช้ Default ของ Prometheus ที่รวม Go Metrics ไว้แล้ว
    prome.RegisterAt(app, "/metrics")

    // Swagger route
    app.Get("/swagger/*", swagger.HandlerDefault)

    // 3. Routes
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.SendString("OK")
    })

    api := app.Group("/api/v1/kitchen")
    api.Post("/tickets", hdl.CreateTicket)       // รับ Order จาก Order Service
    api.Patch("/tickets/:orderId", hdl.UpdateStatus) // อัปเดตเมื่อทำเสร็จ
    api.Get("/orders/:orderId", hdl.GetStatus)      // เช็คสถานะ (KrakenD map จาก /status/:id)

    // 4. Register with Consul
    registerWithConsul()

    app.Listen(":3001")
}

func registerWithConsul() {
    config := consul.DefaultConfig()
    client, err := consul.NewClient(config)
    if err != nil {
        log.Error().Err(err).Msg("Error creating consul client")
        return
    }

    address := "localhost"
    if envAddr := os.Getenv("SERVICE_ADDRESS"); envAddr != "" {
        address = envAddr
    }

    registration := &consul.AgentServiceRegistration{
        ID:      "kitchen-service-1",
        Name:    "kitchen-service",
        Port:    3001,
        Address: address,
        Check: &consul.AgentServiceCheck{
            HTTP:     "http://" + address + ":3001/health",
            Interval: "10s",
        },
    }
    
    if err := client.Agent().ServiceRegister(registration); err != nil {
        log.Error().Err(err).Msg("Error registering service")
    } else {
        log.Info().Str("service", "kitchen-service").Msg("Successfully registered kitchen-service with Consul")
    }
}
