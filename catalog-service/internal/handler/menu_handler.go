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

func (h *MenuHandler) GetAllMenuItems(c *fiber.Ctx) error {
	items, err := h.menuService.GetAllMenuItems()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "catalog-service").Int("count", len(items)).Msg("Fetched all menu items")
	return c.JSON(fiber.Map{"menu_items": items})
}

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
