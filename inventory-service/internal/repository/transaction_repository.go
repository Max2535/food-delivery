package repository

import (
	"inventory-service/internal/model"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	FindByMaterialID(materialID uint) ([]model.StockTransaction, error)
	FindAll(limit int) ([]model.StockTransaction, error)
	Create(tx *model.StockTransaction) error
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) FindByMaterialID(materialID uint) ([]model.StockTransaction, error) {
	var txs []model.StockTransaction
	err := r.db.Where("raw_material_id = ?", materialID).Order("created_at desc").Find(&txs).Error
	return txs, err
}

func (r *transactionRepository) FindAll(limit int) ([]model.StockTransaction, error) {
	var txs []model.StockTransaction
	err := r.db.Order("created_at desc").Limit(limit).Find(&txs).Error
	return txs, err
}

func (r *transactionRepository) Create(tx *model.StockTransaction) error {
	return r.db.Create(tx).Error
}
