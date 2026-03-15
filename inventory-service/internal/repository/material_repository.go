package repository

import (
	"inventory-service/internal/model"

	"gorm.io/gorm"
)

type MaterialRepository interface {
	FindAll() ([]model.RawMaterial, error)
	FindByID(id uint) (*model.RawMaterial, error)
	FindByCatalogIngredientID(catalogID uint) (*model.RawMaterial, error)
	FindLowStock() ([]model.RawMaterial, error)
	Create(m *model.RawMaterial) error
	Update(m *model.RawMaterial) error
}

type materialRepository struct {
	db *gorm.DB
}

func NewMaterialRepository(db *gorm.DB) MaterialRepository {
	return &materialRepository{db: db}
}

func (r *materialRepository) FindAll() ([]model.RawMaterial, error) {
	var items []model.RawMaterial
	err := r.db.Find(&items).Error
	return items, err
}

func (r *materialRepository) FindByID(id uint) (*model.RawMaterial, error) {
	var item model.RawMaterial
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *materialRepository) FindByCatalogIngredientID(catalogID uint) (*model.RawMaterial, error) {
	var item model.RawMaterial
	if err := r.db.Where("catalog_ingredient_id = ?", catalogID).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *materialRepository) FindLowStock() ([]model.RawMaterial, error) {
	var items []model.RawMaterial
	// current_stock < reorder_point AND reorder_point > 0
	err := r.db.Where("current_stock < reorder_point AND reorder_point > 0").Find(&items).Error
	return items, err
}

func (r *materialRepository) Create(m *model.RawMaterial) error {
	return r.db.Create(m).Error
}

func (r *materialRepository) Update(m *model.RawMaterial) error {
	return r.db.Save(m).Error
}
