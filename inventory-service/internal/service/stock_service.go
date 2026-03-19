package service

import (
	"errors"
	"fmt"

	"inventory-service/internal/catalog"
	"inventory-service/internal/model"
	"inventory-service/internal/repository"

	"github.com/rs/zerolog/log"
)

var ErrInsufficientStock = errors.New("insufficient stock")

// DeductItem describes one menu item in a deduction request.
type DeductItem struct {
	MenuItemID        uint    `json:"menu_item_id"`
	Quantity          int     `json:"quantity"`           // number of portions ordered
	PortionMultiplier float64 `json:"portion_multiplier"` // 1.0 = ธรรมดา, 1.5 = พิเศษ
}

type StockService interface {
	Restock(materialID uint, qty float64, note string, correlationID string) (*model.RawMaterial, *model.StockTransaction, error)
	Adjust(materialID uint, qty float64, note string, correlationID string) (*model.RawMaterial, *model.StockTransaction, error)
	DeductByBOM(orderID *uint, items []DeductItem, correlationID string) error
}

type stockService struct {
	materialRepo    repository.MaterialRepository
	transactionRepo repository.TransactionRepository
	catalogClient   *catalog.Client
}

func NewStockService(
	materialRepo repository.MaterialRepository,
	transactionRepo repository.TransactionRepository,
	catalogClient *catalog.Client,
) StockService {
	return &stockService{
		materialRepo:    materialRepo,
		transactionRepo: transactionRepo,
		catalogClient:   catalogClient,
	}
}

func (s *stockService) Restock(materialID uint, qty float64, note string, correlationID string) (*model.RawMaterial, *model.StockTransaction, error) {
	return s.applyStockChange(materialID, qty, model.TransactionRestock, nil, note, correlationID)
}

func (s *stockService) Adjust(materialID uint, qty float64, note string, correlationID string) (*model.RawMaterial, *model.StockTransaction, error) {
	return s.applyStockChange(materialID, qty, model.TransactionAdjustment, nil, note, correlationID)
}

// DeductByBOM looks up BOM from Catalog Service for each item, then deducts stock.
// Logs a low-stock warning for each material that drops below its reorder point.
func (s *stockService) DeductByBOM(orderID *uint, items []DeductItem, correlationID string) error {
	for _, item := range items {
		if item.PortionMultiplier <= 0 {
			item.PortionMultiplier = 1.0
		}

		bomItems, err := s.catalogClient.GetBOM(item.MenuItemID)
		if err != nil {
			log.Warn().Err(err).
				Str("correlation_id", correlationID).
				Uint("menu_item_id", item.MenuItemID).
				Msg("Could not fetch BOM from catalog; skipping deduction for this item")
			continue
		}

		for _, bom := range bomItems {
			material, err := s.materialRepo.FindByCatalogIngredientID(bom.IngredientID)
			if err != nil {
				log.Warn().
					Str("correlation_id", correlationID).
					Uint("catalog_ingredient_id", bom.IngredientID).
					Msg("No raw material linked to catalog ingredient; skipping")
				continue
			}

			deductQty := bom.Quantity * float64(item.Quantity) * item.PortionMultiplier
			note := fmt.Sprintf("auto-deduct: menu_item_id=%d qty=%d portion=%.2f", item.MenuItemID, item.Quantity, item.PortionMultiplier)

			if _, _, err := s.applyStockChange(material.ID, -deductQty, model.TransactionDeduction, orderID, note, correlationID); err != nil {
				log.Error().Err(err).
					Str("correlation_id", correlationID).
					Uint("material_id", material.ID).
					Msg("Failed to deduct stock")
			}
		}
	}
	return nil
}

// applyStockChange updates CurrentStock and records a StockTransaction atomically.
func (s *stockService) applyStockChange(materialID uint, delta float64, txType string, orderID *uint, note string, correlationID string) (*model.RawMaterial, *model.StockTransaction, error) {
	material, err := s.materialRepo.FindByID(materialID)
	if err != nil {
		return nil, nil, ErrMaterialNotFound
	}

	material.CurrentStock += delta

	if err := s.materialRepo.Update(material); err != nil {
		return nil, nil, err
	}

	tx := &model.StockTransaction{
		RawMaterialID:  materialID,
		QuantityChange: delta,
		Type:           txType,
		OrderID:        orderID,
		CorrelationID:  correlationID,
		Note:           note,
	}
	if err := s.transactionRepo.Create(tx); err != nil {
		log.Error().Err(err).Str("correlation_id", correlationID).Msg("Failed to record stock transaction")
	}

	// Low-stock alert
	if material.ReorderPoint > 0 && material.CurrentStock < material.ReorderPoint {
		log.Warn().
			Str("service", "inventory-service").
			Str("alert", "LOW_STOCK").
			Uint("raw_material_id", material.ID).
			Str("name", material.Name).
			Float64("current_stock", material.CurrentStock).
			Float64("reorder_point", material.ReorderPoint).
			Msg("Stock below reorder point")
	}

	return material, tx, nil
}
