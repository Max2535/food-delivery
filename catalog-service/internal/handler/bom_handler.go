package handler

import (
	"errors"
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
	IngredientID  *uint   `json:"ingredient_id"`
	SubMenuItemID *uint   `json:"sub_menu_item_id"`
	Quantity      float64 `json:"quantity"`
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

// GetFlatBOM recursively expands all sub-recipes and returns a flat list of raw ingredients.
// This endpoint is used by inventory-service for stock deduction.
func (h *BOMHandler) GetFlatBOM(c *fiber.Ctx) error {
	menuItemID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid menu item ID"})
	}
	items, err := h.service.GetFlatBOM(uint(menuItemID))
	if err != nil {
		if errors.Is(err, service.ErrCircularBOM) {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("menu_item_id", menuItemID).Msg("Fetched flat BOM")
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
	if req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "quantity must be positive"})
	}

	item, err := h.service.AddBOMItem(uint(menuItemID), req.IngredientID, req.SubMenuItemID, req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidBOMEntry):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, service.ErrBOMIngredientNotFound), errors.Is(err, service.ErrBOMMenuItemNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, service.ErrCircularBOM):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("menu_item_id", menuItemID).Msg("BOM item added")
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
