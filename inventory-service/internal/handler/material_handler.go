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
// GetAll godoc
// @Summary      Get all raw materials
// @Description  Get a list of all raw materials in inventory
// @Tags         materials
// @Produce      json
// @Success      200  {object}  map[string][]model.RawMaterial
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/inventory/materials [get]
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
// GetLowStock godoc
// @Summary      Get low-stock materials
// @Description  Get a list of raw materials that are below their reorder point
// @Tags         materials
// @Produce      json
// @Success      200  {object}  map[string][]model.RawMaterial
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/inventory/materials/low-stock [get]
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
// GetByID godoc
// @Summary      Get material by ID
// @Description  Get detailed information about a specific raw material
// @Tags         materials
// @Produce      json
// @Param        id   path      int  true  "Material ID"
// @Success      200  {object}  model.RawMaterial
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/inventory/materials/{id} [get]
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
// Create godoc
// @Summary      Create a raw material
// @Description  Add a new raw material to the inventory
// @Tags         materials
// @Accept       json
// @Produce      json
// @Param        material body      model.RawMaterial true  "Material Data"
// @Success      201      {object}  model.RawMaterial
// @Failure      400      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /api/v1/inventory/materials [post]
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
// Update godoc
// @Summary      Update a raw material
// @Description  Update an existing raw material by its ID
// @Tags         materials
// @Accept       json
// @Produce      json
// @Param        id       path      int                true  "Material ID"
// @Param        material body      model.RawMaterial  true  "Updated Material Data"
// @Success      200      {object}  model.RawMaterial
// @Failure      400      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Router       /api/v1/inventory/materials/{id} [put]
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
