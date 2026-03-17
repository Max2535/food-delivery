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

// GetAllIngredients godoc
// @Summary      Get all ingredients
// @Description  Get a list of all ingredients
// @Tags         ingredients
// @Produce      json
// @Success      200  {object}  map[string][]model.Ingredient
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/catalog/ingredients [get]
func (h *IngredientHandler) GetAllIngredients(c *fiber.Ctx) error {
	ingredients, err := h.service.GetAllIngredients()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("count", len(ingredients)).Msg("Fetched all ingredients")
	return c.JSON(fiber.Map{"ingredients": ingredients})
}

// CreateIngredient godoc
// @Summary      Create a new ingredient
// @Description  Create a new ingredient with the provided data
// @Tags         ingredients
// @Accept       json
// @Produce      json
// @Param        ingredient body model.Ingredient true "Ingredient Data"
// @Success      201  {object}  model.Ingredient
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/catalog/ingredients [post]
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
