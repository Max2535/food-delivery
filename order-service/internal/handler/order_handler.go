package handler

import (
	"fmt"
	"strconv"

	"order-service/internal/model"
	"order-service/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type OrderHandler struct {
	service service.OrderService
}

func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// CreateOrder godoc
// @Summary      Create a new order
// @Description  Create a new order with items (menu snapshot pattern)
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        order body model.CreateOrderRequest true "Order Data with Items"
// @Success      201  {object}  model.Order
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	req := new(model.CreateOrderRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Validate — ต้องมี customer_id และ items
	if req.CustomerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "customer_id is required"})
	}
	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "at least one item is required"})
	}

	correlationID, _ := c.Locals("correlationID").(string)

	order, err := h.service.CreateOrderFromRequest(req, correlationID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	log.Info().
		Str("service", "order-service").
		Str("order_id", fmt.Sprint(order.ID)).
		Str("correlation_id", correlationID).
		Int("item_count", len(order.Items)).
		Msg("Order created successfully")

	return c.Status(fiber.StatusCreated).JSON(order)
}

// GetAllOrders godoc
// @Summary      Get all orders
// @Description  Get a list of all orders with their items
// @Tags         orders
// @Produce      json
// @Success      200  {array}   model.Order
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/orders [get]
func (h *OrderHandler) GetAllOrders(c *fiber.Ctx) error {
	orders, err := h.service.GetAllOrders()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all orders")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().
		Str("service", "order-service").
		Str("correlation_id", correlationID).
		Int("count", len(orders)).
		Msg("Fetched all orders")

	return c.JSON(fiber.Map{"orders": orders})
}

// GetOrderByID godoc
// @Summary      Get an order by ID
// @Description  Get detailed information about a specific order by its ID, including items
// @Tags         orders
// @Produce      json
// @Param        id   path      int  true  "Order ID"
// @Success      200  {object}  model.Order
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/orders/{id} [get]
func (h *OrderHandler) GetOrderByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		log.Warn().Err(err).Str("id", c.Params("id")).Msg("Invalid order ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	order, err := h.service.GetOrderByID(uint(id))
	if err != nil {
		log.Warn().Err(err).Int("id", id).Msg("Order not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}

	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().
		Str("service", "order-service").
		Str("correlation_id", correlationID).
		Int("order_id", id).
		Msg("Fetched order by ID")

	return c.JSON(order)
}
