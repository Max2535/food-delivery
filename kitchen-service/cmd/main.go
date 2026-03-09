package main

import (
    "kitchen-service/internal/handler"
    "kitchen-service/internal/repository"
    "kitchen-service/internal/service"
    "github.com/gofiber/fiber/v2"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "os"
)

func main() {
    // 1. Database Connection (Kitchen DB)
    dsn := os.Getenv("DB_URL") 
    db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    
    // 2. Setup Layers
    repo := repository.NewKitchenRepository(db)
    svc := service.NewKitchenService(repo)
    hdl := handler.NewKitchenHandler(svc)

    app := fiber.New()

    // 3. Routes
    api := app.Group("/api/v1/kitchen")
    api.Post("/tickets", hdl.CreateTicket)       // รับ Order จาก Order Service
    api.Patch("/tickets/:orderId", hdl.UpdateStatus) // อัปเดตเมื่อทำเสร็จ

    app.Listen(":3001")
}
