package repository

import (
	"catalog-service/internal/model"

	"gorm.io/gorm"
)

type PortionRepository interface {
	FindByMenuItemID(menuItemID uint) ([]model.MenuPortionSize, error)
	Create(portion *model.MenuPortionSize) error
	Delete(id uint) error
}

type portionRepository struct {
	db *gorm.DB
}

func NewPortionRepository(db *gorm.DB) PortionRepository {
	return &portionRepository{db: db}
}

func (r *portionRepository) FindByMenuItemID(menuItemID uint) ([]model.MenuPortionSize, error) {
	var portions []model.MenuPortionSize
	err := r.db.Where("menu_item_id = ?", menuItemID).Find(&portions).Error
	return portions, err
}

func (r *portionRepository) Create(portion *model.MenuPortionSize) error {
	return r.db.Create(portion).Error
}

func (r *portionRepository) Delete(id uint) error {
	return r.db.Delete(&model.MenuPortionSize{}, id).Error
}
