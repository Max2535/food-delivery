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
