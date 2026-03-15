package repository

import (
	"catalog-service/internal/model"

	"gorm.io/gorm"
)

type MenuRepository interface {
	FindAll() ([]model.MenuItem, error)
	FindByID(id uint) (*model.MenuItem, error)
	Create(item *model.MenuItem) error
	Update(item *model.MenuItem) error
	Delete(id uint) error
	FindByName(name string) (*model.MenuItem, error)
}

type menuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &menuRepository{db: db}
}

func (r *menuRepository) FindAll() ([]model.MenuItem, error) {
	var items []model.MenuItem
	err := r.db.Find(&items).Error
	return items, err
}

func (r *menuRepository) FindByID(id uint) (*model.MenuItem, error) {
	var item model.MenuItem
	err := r.db.First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *menuRepository) Create(item *model.MenuItem) error {
	return r.db.Create(item).Error
}

func (r *menuRepository) Update(item *model.MenuItem) error {
	return r.db.Save(item).Error
}

func (r *menuRepository) Delete(id uint) error {
	return r.db.Delete(&model.MenuItem{}, id).Error
}

func (r *menuRepository) FindByName(name string) (*model.MenuItem, error) {
	var item model.MenuItem
	err := r.db.Where("name = ?", name).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
