package handler

import (
	"strconv"

	"inventory-service/internal/model"
	"inventory-service/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type MaterialHandler struct {
	service service.MaterialService
}

func NewMaterialHandler(s service.MaterialService) *MaterialHandler {
	return &MaterialHandler{service: s}
}

// GET /api/v1/inventory/materials
func (h *MaterialHandler) GetAll(c *fiber.Ctx) error {
	items, err := h.service.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "inventory-service").Int("count", len(items)).Msg("Fetched all raw materials")
	return c.JSON(fiber.Map{"raw_materials": items})
}

// GET /api/v1/inventory/materials/low-stock
func (h *MaterialHandler) GetLowStock(c *fiber.Ctx) error {
	items, err := h.service.GetLowStock()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "inventory-service").Int("count", len(items)).Msg("Fetched low-stock materials")
	return c.JSON(fiber.Map{"low_stock": items})
}

// GET /api/v1/inventory/materials/:id
func (h *MaterialHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}
	item, err := h.service.GetByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(item)
}

// POST /api/v1/inventory/materials
func (h *MaterialHandler) Create(c *fiber.Ctx) error {
	m := new(model.RawMaterial)
	if err := c.BodyParser(m); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if m.Name == "" || m.Unit == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name and unit are required"})
	}
	if err := h.service.Create(m); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "inventory-service").Uint("material_id", m.ID).Msg("Raw material created")
	return c.Status(fiber.StatusCreated).JSON(m)
}

// PUT /api/v1/inventory/materials/:id
func (h *MaterialHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}
	input := new(model.RawMaterial)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	updated, err := h.service.Update(uint(id), input)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "inventory-service").Uint("material_id", updated.ID).Msg("Raw material updated")
	return c.JSON(updated)
}
