package handler

import (
	"fmt"
	"kitchen-service/internal/model"
	"kitchen-service/internal/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type KitchenHandler struct {
	service service.KitchenService
}

func NewKitchenHandler(service service.KitchenService) *KitchenHandler {
	return &KitchenHandler{service: service}
}

// CreateTicket รับข้อมูลจาก Order Service เพื่อสร้างใบสั่งงานในครัว
func (h *KitchenHandler) CreateTicket(c *fiber.Ctx) error {
	ticket := new(model.KitchenTicket)
	if err := c.BodyParser(ticket); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ไม่สามารถอ่านข้อมูล Ticket ได้",
		})
	}

	if err := h.service.CreateTicket(ticket); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถสร้างรายการในครัวได้",
		})
	}

	correlationID := c.Get("X-Correlation-ID")
	if correlationID == "" {
		correlationID = "unknown"
	}

	log.Info().
		Str("service", "kitchen-service").
		Str("order_id", fmt.Sprint(ticket.OrderID)).
		Str("correlation_id", correlationID). // สำคัญมากสำหรับการแกะรอย
		Msg("Ticket created successfully")

	return c.Status(fiber.StatusCreated).JSON(ticket)
}

// UpdateStatus ใช้สำหรับให้กุ๊กอัปเดตสถานะ (เช่น เปลี่ยนเป็น Ready)
func (h *KitchenHandler) UpdateStatus(c *fiber.Ctx) error {
	// รับ OrderID จาก URL parameter
	orderIDStr := c.Params("orderId")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "OrderID ไม่ถูกต้อง",
		})
	}

	// รับ Status ใหม่จาก JSON body
	type UpdateRequest struct {
		Status string `json:"status"`
	}
	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลสถานะไม่ถูกต้อง",
		})
	}

	// เรียก Service เพื่ออัปเดต DB และส่ง Event ไป RabbitMQ (ถ้า status == Ready)
	if err := h.service.UpdateStatus(uint(orderID), req.Status); err != nil {
		log.Error().Err(err).Uint64("order_id", orderID).Str("status", req.Status).Msg("Failed to update ticket status")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถอัปเดตสถานะได้",
		})
	}

	correlationID := c.Get("X-Correlation-ID", "unknown")
	log.Info().
		Str("service", "kitchen-service").
		Str("correlation_id", correlationID).
		Uint64("order_id", orderID).
		Str("status", req.Status).
		Msg("Ticket status updated")

	return c.JSON(fiber.Map{
		"message": "อัปเดตสถานะเรียบร้อยแล้ว",
		"order_id": orderID,
		"status":   req.Status,
	})
}
