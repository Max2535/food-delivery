package handler

import (
	"strconv"

	"catalog-service/internal/model"
	"catalog-service/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type AddOnHandler struct {
	service service.AddOnService
}

func NewAddOnHandler(service service.AddOnService) *AddOnHandler {
	return &AddOnHandler{service: service}
}

// GET /api/v1/catalog/menus/:id/addons
// GetAddOns godoc
// @Summary      Get add-ons by Menu Item ID
// @Description  Get the optional extras for a specific menu item
// @Tags         addons
// @Produce      json
// @Param        id   path      int  true  "Menu Item ID"
// @Success      200  {object}  map[string][]model.MenuAddOn
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/catalog/menus/{id}/addons [get]
func (h *AddOnHandler) GetAddOns(c *fiber.Ctx) error {
	menuItemID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid menu item ID"})
	}
	addons, err := h.service.GetAddOnsByMenuItemID(uint(menuItemID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("menu_item_id", menuItemID).Msg("Fetched add-ons")
	return c.JSON(fiber.Map{"addons": addons})
}

// POST /api/v1/catalog/menus/:id/addons
// CreateAddOn godoc
// @Summary      Create an add-on
// @Description  Create a new optional extra for a menu item
// @Tags         addons
// @Accept       json
// @Produce      json
// @Param        id    path      int              true  "Menu Item ID"
// @Param        addon body      model.MenuAddOn  true  "Add-on Data"
// @Success      201   {object}  model.MenuAddOn
// @Failure      400   {object}  map[string]interface{}
// @Router       /api/v1/catalog/menus/{id}/addons [post]
func (h *AddOnHandler) CreateAddOn(c *fiber.Ctx) error {
	menuItemID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid menu item ID"})
	}
	addon := new(model.MenuAddOn)
	if err := c.BodyParser(addon); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if addon.IngredientID == 0 || addon.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ingredient_id and quantity are required"})
	}
	addon.MenuItemID = uint(menuItemID)
	if err := h.service.CreateAddOn(addon); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Uint("addon_id", addon.ID).Msg("Add-on created")
	return c.Status(fiber.StatusCreated).JSON(addon)
}

// DELETE /api/v1/catalog/menus/:id/addons/:addon_id
// DeleteAddOn godoc
// @Summary      Delete an add-on
// @Description  Delete an optional extra by its ID
// @Tags         addons
// @Param        addon_id path      int  true  "Add-on ID"
// @Success      204      "No Content"
// @Failure      400      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Router       /api/v1/catalog/addons/{addon_id} [delete]
func (h *AddOnHandler) DeleteAddOn(c *fiber.Ctx) error {
	addonID, err := strconv.Atoi(c.Params("addon_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid add-on ID"})
	}
	if err := h.service.DeleteAddOn(uint(addonID)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Add-on not found"})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("addon_id", addonID).Msg("Add-on deleted")
	return c.SendStatus(fiber.StatusNoContent)
}
