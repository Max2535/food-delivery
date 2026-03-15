package handler

import (
	"strconv"

	"catalog-service/internal/model"
	"catalog-service/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type PortionHandler struct {
	service service.PortionService
}

func NewPortionHandler(service service.PortionService) *PortionHandler {
	return &PortionHandler{service: service}
}

// GET /api/v1/catalog/menus/:id/portions
func (h *PortionHandler) GetPortions(c *fiber.Ctx) error {
	menuItemID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid menu item ID"})
	}
	portions, err := h.service.GetPortionsByMenuItemID(uint(menuItemID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("menu_item_id", menuItemID).Msg("Fetched portions")
	return c.JSON(fiber.Map{"portions": portions})
}

// POST /api/v1/catalog/menus/:id/portions
func (h *PortionHandler) CreatePortion(c *fiber.Ctx) error {
	menuItemID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid menu item ID"})
	}
	portion := new(model.MenuPortionSize)
	if err := c.BodyParser(portion); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if portion.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}
	if portion.QuantityMultiplier == 0 {
		portion.QuantityMultiplier = 1.0
	}
	portion.MenuItemID = uint(menuItemID)
	if err := h.service.CreatePortion(portion); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Uint("portion_id", portion.ID).Msg("Portion size created")
	return c.Status(fiber.StatusCreated).JSON(portion)
}

// DELETE /api/v1/catalog/menus/:id/portions/:portion_id
func (h *PortionHandler) DeletePortion(c *fiber.Ctx) error {
	portionID, err := strconv.Atoi(c.Params("portion_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid portion ID"})
	}
	if err := h.service.DeletePortion(uint(portionID)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Portion not found"})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("portion_id", portionID).Msg("Portion size deleted")
	return c.SendStatus(fiber.StatusNoContent)
}
