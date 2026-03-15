package handler

import (
	"catalog-service/internal/model"
	"catalog-service/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type IngredientHandler struct {
	service service.IngredientService
}

func NewIngredientHandler(service service.IngredientService) *IngredientHandler {
	return &IngredientHandler{service: service}
}

func (h *IngredientHandler) GetAllIngredients(c *fiber.Ctx) error {
	ingredients, err := h.service.GetAllIngredients()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("count", len(ingredients)).Msg("Fetched all ingredients")
	return c.JSON(fiber.Map{"ingredients": ingredients})
}

func (h *IngredientHandler) CreateIngredient(c *fiber.Ctx) error {
	ingredient := new(model.Ingredient)
	if err := c.BodyParser(ingredient); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if ingredient.Name == "" || ingredient.Unit == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name and unit are required"})
	}
	if err := h.service.CreateIngredient(ingredient); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Uint("ingredient_id", ingredient.ID).Msg("Ingredient created")
	return c.Status(fiber.StatusCreated).JSON(ingredient)
}
