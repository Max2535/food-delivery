package handler

import (
	"strconv"

	"catalog-service/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type BOMHandler struct {
	service service.BOMService
}

func NewBOMHandler(service service.BOMService) *BOMHandler {
	return &BOMHandler{service: service}
}

type addBOMRequest struct {
	IngredientID uint    `json:"ingredient_id"`
	Quantity     float64 `json:"quantity"`
}

func (h *BOMHandler) GetBOM(c *fiber.Ctx) error {
	menuItemID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid menu item ID"})
	}
	items, err := h.service.GetBOMByMenuItemID(uint(menuItemID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("menu_item_id", menuItemID).Msg("Fetched BOM")
	return c.JSON(fiber.Map{"bom_items": items})
}

func (h *BOMHandler) AddBOMItem(c *fiber.Ctx) error {
	menuItemID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid menu item ID"})
	}
	var req addBOMRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if req.IngredientID == 0 || req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ingredient_id and quantity are required"})
	}
	item, err := h.service.AddBOMItem(uint(menuItemID), req.IngredientID, req.Quantity)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("menu_item_id", menuItemID).Uint("ingredient_id", req.IngredientID).Msg("BOM item added")
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *BOMHandler) DeleteBOMItem(c *fiber.Ctx) error {
	bomID, err := strconv.Atoi(c.Params("bom_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid BOM item ID"})
	}
	if err := h.service.DeleteBOMItem(uint(bomID)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "BOM item not found"})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("bom_id", bomID).Msg("BOM item deleted")
	return c.SendStatus(fiber.StatusNoContent)
}
