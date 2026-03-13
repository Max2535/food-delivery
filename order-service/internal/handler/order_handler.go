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
// @Description  Create a new order based on provided data
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        order body model.Order true "Order Data"
// @Success      201  {object}  model.Order
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	order := new(model.Order)
	if err := c.BodyParser(order); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if err := h.service.CreateOrder(order); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	correlationID := c.Get("X-Correlation-ID")
	if correlationID == "" {
		correlationID = "unknown"
	}

	log.Info().
		Str("service", "order-service").
		Str("order_id", fmt.Sprint(order.ID)).
		Str("correlation_id", correlationID). // สำคัญมากสำหรับการแกะรอย
		Msg("Order created successfully")

	return c.Status(fiber.StatusCreated).JSON(order)
}

// GetAllOrders godoc
// @Summary      Get all orders
// @Description  Get a list of all orders
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

	correlationID := c.Get("X-Correlation-ID", "unknown")
	log.Info().
		Str("service", "order-service").
		Str("correlation_id", correlationID).
		Int("count", len(orders)).
		Msg("Fetched all orders")

	return c.JSON(orders)
}

// GetOrderByID godoc
// @Summary      Get an order by ID
// @Description  Get detailed information about a specific order by its ID
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

	correlationID := c.Get("X-Correlation-ID", "unknown")
	log.Info().
		Str("service", "order-service").
		Str("correlation_id", correlationID).
		Int("order_id", id).
		Msg("Fetched order by ID")

	return c.JSON(order)
}
