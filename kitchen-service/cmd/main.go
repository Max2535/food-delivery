package main

import (
    "kitchen-service/internal/handler"
    "kitchen-service/internal/model"
    "kitchen-service/internal/repository"
    "kitchen-service/internal/service"
    "kitchen-service/internal/worker"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	consul "github.com/hashicorp/consul/api"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found")
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

    // 3. Routes
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.SendString("OK")
    })

    api := app.Group("/api/v1/kitchen")
    api.Post("/tickets", hdl.CreateTicket)       // รับ Order จาก Order Service
    api.Patch("/tickets/:orderId", hdl.UpdateStatus) // อัปเดตเมื่อทำเสร็จ

    // 4. Register with Consul
    registerWithConsul()

    app.Listen(":3001")
}

func registerWithConsul() {
    config := consul.DefaultConfig()
    client, err := consul.NewClient(config)
    if err != nil {
        log.Printf("Error creating consul client: %v", err)
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
        log.Printf("Error registering service: %v", err)
    } else {
        log.Println("Successfully registered kitchen-service with Consul")
    }
}
