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

// GetBOM godoc
// @Summary      Get BOM by Menu Item ID
// @Description  Get the Bill of Materials for a specific menu item
// @Tags         bom
// @Produce      json
// @Param        id   path      int  true  "Menu Item ID"
// @Success      200  {object}  map[string][]model.BOMItem
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/catalog/menus/{id}/bom [get]
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
// GetFlatBOM godoc
// @Summary      Get Flat BOM by Menu Item ID
// @Description  Recursively expands all sub-recipes and returns a flat list of raw ingredients
// @Tags         bom
// @Produce      json
// @Param        id   path      int  true  "Menu Item ID"
// @Success      200  {object}  map[string][]service.FlatBOMItem
// @Failure      400  {object}  map[string]interface{}
// @Failure      422  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/catalog/menus/{id}/bom/flat [get]
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

// AddBOMItem godoc
// @Summary      Add a BOM item
// @Description  Add an ingredient or sub-menu item to a menu item's BOM
// @Tags         bom
// @Accept       json
// @Produce      json
// @Param        id   path      int           true  "Menu Item ID"
// @Param        req  body      addBOMRequest true  "BOM Item Data"
// @Success      201  {object}  model.BOMItem
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      422  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/catalog/menus/{id}/bom [post]
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

// DeleteBOMItem godoc
// @Summary      Delete a BOM item
// @Description  Delete a BOM item by its ID
// @Tags         bom
// @Param        bom_id path      int  true  "BOM Item ID"
// @Success      204    "No Content"
// @Failure      400    {object}  map[string]interface{}
// @Failure      404    {object}  map[string]interface{}
// @Router       /api/v1/catalog/bom/{bom_id} [delete]
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
