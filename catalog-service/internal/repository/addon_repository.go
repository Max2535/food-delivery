package repository

import (
	"catalog-service/internal/model"

	"gorm.io/gorm"
)

type AddOnRepository interface {
	FindByMenuItemID(menuItemID uint) ([]model.MenuAddOn, error)
	Create(addon *model.MenuAddOn) error
	Delete(id uint) error
}

type addOnRepository struct {
	db *gorm.DB
}

func NewAddOnRepository(db *gorm.DB) AddOnRepository {
	return &addOnRepository{db: db}
}

func (r *addOnRepository) FindByMenuItemID(menuItemID uint) ([]model.MenuAddOn, error) {
	var addons []model.MenuAddOn
	err := r.db.Preload("Ingredient").Where("menu_item_id = ?", menuItemID).Find(&addons).Error
	return addons, err
}

func (r *addOnRepository) Create(addon *model.MenuAddOn) error {
	return r.db.Create(addon).Error
}

func (r *addOnRepository) Delete(id uint) error {
	return r.db.Delete(&model.MenuAddOn{}, id).Error
}
