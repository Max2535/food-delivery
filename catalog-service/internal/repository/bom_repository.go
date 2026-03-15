package repository

import (
	"catalog-service/internal/model"

	"gorm.io/gorm"
)

type BOMRepository interface {
	FindByMenuItemID(menuItemID uint) ([]model.BOMItem, error)
	AddItem(item *model.BOMItem) error
	DeleteItem(id uint) error
}

type bomRepository struct {
	db *gorm.DB
}

func NewBOMRepository(db *gorm.DB) BOMRepository {
	return &bomRepository{db: db}
}

func (r *bomRepository) FindByMenuItemID(menuItemID uint) ([]model.BOMItem, error) {
	var items []model.BOMItem
	err := r.db.Preload("Ingredient").Where("menu_item_id = ?", menuItemID).Find(&items).Error
	return items, err
}

func (r *bomRepository) AddItem(item *model.BOMItem) error {
	return r.db.Create(item).Error
}

func (r *bomRepository) DeleteItem(id uint) error {
	return r.db.Delete(&model.BOMItem{}, id).Error
}
