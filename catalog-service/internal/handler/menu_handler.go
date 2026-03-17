package handler

import (
	"strconv"

	"catalog-service/internal/model"
	"catalog-service/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type MenuHandler struct {
	menuService    service.MenuService
	stationService service.StationService
}

func NewMenuHandler(menuService service.MenuService, stationService service.StationService) *MenuHandler {
	return &MenuHandler{menuService: menuService, stationService: stationService}
}

// GetAllMenuItems godoc
// @Summary      Get all menu items
// @Description  Get a list of all menu items
// @Tags         menus
// @Produce      json
// @Success      200  {object}  map[string][]model.MenuItem
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/catalog/menus [get]
func (h *MenuHandler) GetAllMenuItems(c *fiber.Ctx) error {
	items, err := h.menuService.GetAllMenuItems()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("count", len(items)).Msg("Fetched all menu items")
	return c.JSON(fiber.Map{"menu_items": items})
}

// GetMenuItemByID godoc
// @Summary      Get a menu item by ID
// @Description  Get detailed information about a specific menu item by its ID
// @Tags         menus
// @Produce      json
// @Param        id   path      int  true  "Menu Item ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/catalog/menus/{id} [get]
func (h *MenuHandler) GetMenuItemByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	item, err := h.menuService.GetMenuItemByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Menu item not found"})
	}

	// Enrich with station info
	stations, _ := h.stationService.GetStationsByMenuItemID(uint(id))

	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Uint("menu_item_id", item.ID).Msg("Fetched menu item by ID")

	return c.JSON(fiber.Map{
		"menu_item": item,
		"stations":  stations,
	})
}

// CreateMenuItem godoc
// @Summary      Create a new menu item
// @Description  Create a new menu item with the provided data
// @Tags         menus
// @Accept       json
// @Produce      json
// @Param        menu body model.MenuItem true "Menu Item Data"
// @Success      201  {object}  model.MenuItem
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/catalog/menus [post]
func (h *MenuHandler) CreateMenuItem(c *fiber.Ctx) error {
	item := new(model.MenuItem)
	if err := c.BodyParser(item); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if item.Name == "" || item.Price <= 0 || item.Category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name, price, and category are required"})
	}
	if err := h.menuService.CreateMenuItem(item); err != nil {
		if err == service.ErrMenuItemAlreadyExists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Uint("menu_item_id", item.ID).Msg("Menu item created")
	return c.Status(fiber.StatusCreated).JSON(item)
}

// UpdateMenuItem godoc
// @Summary      Update a menu item
// @Description  Update an existing menu item by its ID
// @Tags         menus
// @Accept       json
// @Produce      json
// @Param        id   path      int            true  "Menu Item ID"
// @Param        menu body      model.MenuItem true "Updated Menu Item Data"
// @Success      200  {object}  model.MenuItem
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/catalog/menus/{id} [put]
func (h *MenuHandler) UpdateMenuItem(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}
	input := new(model.MenuItem)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	updated, err := h.menuService.UpdateMenuItem(uint(id), input)
	if err != nil {
		if err == service.ErrMenuItemAlreadyExists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Menu item not found"})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Uint("menu_item_id", updated.ID).Msg("Menu item updated")
	return c.JSON(updated)
}

// DeleteMenuItem godoc
// @Summary      Delete a menu item
// @Description  Delete a menu item by its ID
// @Tags         menus
// @Param        id   path      int  true  "Menu Item ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/catalog/menus/{id} [delete]
func (h *MenuHandler) DeleteMenuItem(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}
	if err := h.menuService.DeleteMenuItem(uint(id)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Menu item not found"})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("menu_item_id", id).Msg("Menu item deleted")
	return c.SendStatus(fiber.StatusNoContent)
}
