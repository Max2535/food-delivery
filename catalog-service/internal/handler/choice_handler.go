package handler

import (
	"strconv"

	"catalog-service/internal/model"
	"catalog-service/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type ChoiceHandler struct {
	service service.ChoiceService
}

func NewChoiceHandler(service service.ChoiceService) *ChoiceHandler {
	return &ChoiceHandler{service: service}
}

// GET /api/v1/catalog/menus/:id/choices
// GetChoices godoc
// @Summary      Get choices by Menu Item ID
// @Description  Get the Bill of Materials choice groups for a specific menu item
// @Tags         choices
// @Produce      json
// @Param        id   path      int  true  "Menu Item ID"
// @Success      200  {object}  map[string][]model.BOMChoiceGroup
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/catalog/menus/{id}/choices [get]
func (h *ChoiceHandler) GetChoices(c *fiber.Ctx) error {
	menuItemID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid menu item ID"})
	}
	groups, err := h.service.GetChoicesByMenuItemID(uint(menuItemID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("menu_item_id", menuItemID).Msg("Fetched choice groups")
	return c.JSON(fiber.Map{"choice_groups": groups})
}

// POST /api/v1/catalog/menus/:id/choices
// CreateChoiceGroup godoc
// @Summary      Create a choice group
// @Description  Create a new BOM choice group for a menu item
// @Tags         choices
// @Accept       json
// @Produce      json
// @Param        id    path      int                     true  "Menu Item ID"
// @Param        group body      model.BOMChoiceGroup    true  "Choice Group Data"
// @Success      201   {object}  model.BOMChoiceGroup
// @Failure      400   {object}  map[string]interface{}
// @Router       /api/v1/catalog/menus/{id}/choices [post]
func (h *ChoiceHandler) CreateChoiceGroup(c *fiber.Ctx) error {
	menuItemID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid menu item ID"})
	}
	group := new(model.BOMChoiceGroup)
	if err := c.BodyParser(group); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if group.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}
	group.MenuItemID = uint(menuItemID)
	if group.MinChoices == 0 {
		group.MinChoices = 1
	}
	if group.MaxChoices == 0 {
		group.MaxChoices = 1
	}
	if err := h.service.CreateChoiceGroup(group); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Uint("group_id", group.ID).Msg("Choice group created")
	return c.Status(fiber.StatusCreated).JSON(group)
}

// DELETE /api/v1/catalog/menus/:id/choices/:group_id
// DeleteChoiceGroup godoc
// @Summary      Delete a choice group
// @Description  Delete a BOM choice group by its ID
// @Tags         choices
// @Param        group_id path      int  true  "Group ID"
// @Success      204    "No Content"
// @Failure      400    {object}  map[string]interface{}
// @Failure      404    {object}  map[string]interface{}
// @Router       /api/v1/catalog/choices/{group_id} [delete]
func (h *ChoiceHandler) DeleteChoiceGroup(c *fiber.Ctx) error {
	groupID, err := strconv.Atoi(c.Params("group_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}
	if err := h.service.DeleteChoiceGroup(uint(groupID)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("group_id", groupID).Msg("Choice group deleted")
	return c.SendStatus(fiber.StatusNoContent)
}

type addChoiceOptionRequest struct {
	IngredientID uint    `json:"ingredient_id"`
	Quantity     float64 `json:"quantity"`
	ExtraPrice   float64 `json:"extra_price"`
}

// POST /api/v1/catalog/menus/:id/choices/:group_id/options
// AddChoiceOption godoc
// @Summary      Add a choice option
// @Description  Add an ingredient option to a BOM choice group
// @Tags         choices
// @Accept       json
// @Produce      json
// @Param        group_id path      int                     true  "Group ID"
// @Param        req      body      addChoiceOptionRequest  true  "Option Data"
// @Success      201      {object}  model.BOMChoiceOption
// @Failure      400      {object}  map[string]interface{}
// @Router       /api/v1/catalog/choices/{group_id}/options [post]
func (h *ChoiceHandler) AddChoiceOption(c *fiber.Ctx) error {
	groupID, err := strconv.Atoi(c.Params("group_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}
	var req addChoiceOptionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if req.IngredientID == 0 || req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ingredient_id and quantity are required"})
	}
	option, err := h.service.AddChoiceOption(uint(groupID), req.IngredientID, req.Quantity, req.ExtraPrice)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Uint("option_id", option.ID).Msg("Choice option added")
	return c.Status(fiber.StatusCreated).JSON(option)
}

// DELETE /api/v1/catalog/menus/:id/choices/:group_id/options/:option_id
// DeleteChoiceOption godoc
// @Summary      Delete a choice option
// @Description  Delete a choice option by its ID
// @Tags         choices
// @Param        option_id path      int  true  "Option ID"
// @Success      204       "No Content"
// @Failure      400       {object}  map[string]interface{}
// @Failure      404       {object}  map[string]interface{}
// @Router       /api/v1/catalog/options/{option_id} [delete]
func (h *ChoiceHandler) DeleteChoiceOption(c *fiber.Ctx) error {
	optionID, err := strconv.Atoi(c.Params("option_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid option ID"})
	}
	if err := h.service.DeleteChoiceOption(uint(optionID)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Option not found"})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("option_id", optionID).Msg("Choice option deleted")
	return c.SendStatus(fiber.StatusNoContent)
}
