package handler

import (
	"strconv"

	"inventory-service/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type StockHandler struct {
	service service.StockService
}

func NewStockHandler(s service.StockService) *StockHandler {
	return &StockHandler{service: s}
}

type restockRequest struct {
	MaterialID uint    `json:"material_id"`
	Quantity   float64 `json:"quantity"`
	Note       string  `json:"note"`
}

type adjustRequest struct {
	MaterialID uint    `json:"material_id"`
	Quantity   float64 `json:"quantity"` // positive or negative
	Note       string  `json:"note"`
}

type deductRequest struct {
	OrderID *uint                 `json:"order_id"`
	Items   []service.DeductItem  `json:"items"`
}

// POST /api/v1/inventory/stock/restock
// Restock godoc
// @Summary      Restock material
// @Description  Add stock to a specific raw material
// @Tags         stock
// @Accept       json
// @Produce      json
// @Param        req  body      restockRequest true  "Restock Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /api/v1/inventory/stock/restock [post]
func (h *StockHandler) Restock(c *fiber.Ctx) error {
	var req restockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if req.MaterialID == 0 || req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "material_id and positive quantity are required"})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	material, err := h.service.Restock(req.MaterialID, req.Quantity, req.Note, correlationID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	log.Info().Str("correlation_id", correlationID).Str("service", "inventory-service").
		Uint("material_id", req.MaterialID).Float64("qty", req.Quantity).Msg("Stock restocked")
	return c.JSON(fiber.Map{"message": "restocked", "raw_material": material})
}

// POST /api/v1/inventory/stock/adjust
// Adjust godoc
// @Summary      Adjust stock
// @Description  Manually adjust the stock level of a raw material
// @Tags         stock
// @Accept       json
// @Produce      json
// @Param        req  body      adjustRequest true  "Adjustment Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /api/v1/inventory/stock/adjust [post]
func (h *StockHandler) Adjust(c *fiber.Ctx) error {
	var req adjustRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if req.MaterialID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "material_id is required"})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	material, err := h.service.Adjust(req.MaterialID, req.Quantity, req.Note, correlationID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	log.Info().Str("correlation_id", correlationID).Str("service", "inventory-service").
		Uint("material_id", req.MaterialID).Float64("delta", req.Quantity).Msg("Stock adjusted")
	return c.JSON(fiber.Map{"message": "adjusted", "raw_material": material})
}

// POST /api/v1/inventory/stock/deduct
// Manual deduction endpoint — caller provides menu_item_ids and quantities.
// Inventory service fetches BOM from Catalog and deducts accordingly.
// Deduct godoc
// @Summary      Deduct stock via BOM
// @Description  Deduct stock for multiple menu items based on their BOM from Catalog
// @Tags         stock
// @Accept       json
// @Produce      json
// @Param        req  body      deductRequest true  "Deduction Request"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/inventory/stock/deduct [post]
func (h *StockHandler) Deduct(c *fiber.Ctx) error {
	var req deductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "items are required"})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	if err := h.service.DeductByBOM(req.OrderID, req.Items, correlationID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	log.Info().Str("correlation_id", correlationID).Str("service", "inventory-service").
		Int("item_count", len(req.Items)).Msg("Stock deducted via BOM")
	return c.JSON(fiber.Map{"message": "stock deducted successfully"})
}

// GET /api/v1/inventory/stock/:material_id/transactions
func (h *StockHandler) GetTransactions(c *fiber.Ctx) error {
	_ = strconv.Atoi // used below
	materialID, err := strconv.Atoi(c.Params("material_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid material ID"})
	}
	_ = materialID
	// delegated to TransactionHandler; placeholder kept for route clarity
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "use /transactions/:material_id"})
}
