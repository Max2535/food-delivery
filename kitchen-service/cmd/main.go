package main

import (
    "kitchen-service/internal/handler"
    "kitchen-service/internal/model"
    "kitchen-service/internal/repository"
    "kitchen-service/internal/service"
    "kitchen-service/internal/worker"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
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
    api := app.Group("/api/v1/kitchen")
    api.Post("/tickets", hdl.CreateTicket)       // รับ Order จาก Order Service
    api.Patch("/tickets/:orderId", hdl.UpdateStatus) // อัปเดตเมื่อทำเสร็จ

    app.Listen(":3001")
}
