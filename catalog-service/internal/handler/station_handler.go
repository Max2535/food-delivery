package handler

import (
	"strconv"

	"catalog-service/internal/model"
	"catalog-service/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type StationHandler struct {
	service service.StationService
}

func NewStationHandler(service service.StationService) *StationHandler {
	return &StationHandler{service: service}
}

type assignStationRequest struct {
	KitchenStationID uint `json:"kitchen_station_id"`
}

// GetAllStations godoc
// @Summary      Get all kitchen stations
// @Description  Get a list of all defined kitchen stations
// @Tags         stations
// @Produce      json
// @Success      200  {object}  map[string][]model.KitchenStation
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/catalog/stations [get]
func (h *StationHandler) GetAllStations(c *fiber.Ctx) error {
	stations, err := h.service.GetAllStations()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("count", len(stations)).Msg("Fetched all stations")
	return c.JSON(fiber.Map{"stations": stations})
}

// CreateStation godoc
// @Summary      Create a kitchen station
// @Description  Define a new kitchen station
// @Tags         stations
// @Accept       json
// @Produce      json
// @Param        station body      model.KitchenStation true  "Station Data"
// @Success      201     {object}  model.KitchenStation
// @Failure      400     {object}  map[string]interface{}
// @Failure      500     {object}  map[string]interface{}
// @Router       /api/v1/catalog/stations [post]
func (h *StationHandler) CreateStation(c *fiber.Ctx) error {
	station := new(model.KitchenStation)
	if err := c.BodyParser(station); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if station.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}
	if err := h.service.CreateStation(station); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Uint("station_id", station.ID).Msg("Kitchen station created")
	return c.Status(fiber.StatusCreated).JSON(station)
}

// AssignMenuToStation godoc
// @Summary      Assign menu to station
// @Description  Assign a menu item to be handled by a specific kitchen station
// @Tags         stations
// @Accept       json
// @Produce      json
// @Param        id   path      int                   true  "Menu Item ID"
// @Param        req  body      assignStationRequest  true  "Assignment Data"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]interface{}
// @Router       /api/v1/catalog/menus/{id}/station [post]
func (h *StationHandler) AssignMenuToStation(c *fiber.Ctx) error {
	menuItemID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid menu item ID"})
	}
	var req assignStationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if req.KitchenStationID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "kitchen_station_id is required"})
	}
	if err := h.service.AssignMenuToStation(uint(menuItemID), req.KitchenStationID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("menu_item_id", menuItemID).Uint("station_id", req.KitchenStationID).Msg("Menu assigned to station")
	return c.JSON(fiber.Map{"message": "Menu item assigned to station successfully"})
}
