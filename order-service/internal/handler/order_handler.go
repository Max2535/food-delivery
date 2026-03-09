package handler

import (
	"strconv"

	"order-service/internal/model"
	"order-service/internal/service"

	"github.com/gofiber/fiber/v2"
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	order, err := h.service.GetOrderByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}

	return c.JSON(order)
}
