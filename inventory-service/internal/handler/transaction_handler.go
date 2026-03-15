package handler

import (
	"strconv"

	"inventory-service/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type TransactionHandler struct {
	repo repository.TransactionRepository
}

func NewTransactionHandler(repo repository.TransactionRepository) *TransactionHandler {
	return &TransactionHandler{repo: repo}
}

// GET /api/v1/inventory/transactions?limit=50
func (h *TransactionHandler) GetAll(c *fiber.Ctx) error {
	limit := 50
	if l := c.QueryInt("limit", 0); l > 0 {
		limit = l
	}
	txs, err := h.repo.FindAll(limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "inventory-service").Int("count", len(txs)).Msg("Fetched transactions")
	return c.JSON(fiber.Map{"transactions": txs})
}

// GET /api/v1/inventory/transactions/:material_id
func (h *TransactionHandler) GetByMaterial(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("material_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid material ID"})
	}
	txs, err := h.repo.FindByMaterialID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	correlationID, _ := c.Locals("correlationID").(string)
	log.Info().Str("correlation_id", correlationID).Str("service", "inventory-service").Int("material_id", id).Msg("Fetched material transactions")
	return c.JSON(fiber.Map{"transactions": txs})
}
