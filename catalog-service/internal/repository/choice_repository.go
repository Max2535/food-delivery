package repository

import (
	"catalog-service/internal/model"

	"gorm.io/gorm"
)

type ChoiceRepository interface {
	FindGroupsByMenuItemID(menuItemID uint) ([]model.BOMChoiceGroup, error)
	FindGroupByID(id uint) (*model.BOMChoiceGroup, error)
	CreateGroup(group *model.BOMChoiceGroup) error
	DeleteGroup(id uint) error
	AddOption(option *model.BOMChoiceOption) error
	DeleteOption(id uint) error
}

type choiceRepository struct {
	db *gorm.DB
}

func NewChoiceRepository(db *gorm.DB) ChoiceRepository {
	return &choiceRepository{db: db}
}

func (r *choiceRepository) FindGroupsByMenuItemID(menuItemID uint) ([]model.BOMChoiceGroup, error) {
	var groups []model.BOMChoiceGroup
	err := r.db.Preload("Options.Ingredient").Where("menu_item_id = ?", menuItemID).Find(&groups).Error
	return groups, err
}

func (r *choiceRepository) FindGroupByID(id uint) (*model.BOMChoiceGroup, error) {
	var group model.BOMChoiceGroup
	err := r.db.Preload("Options.Ingredient").First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *choiceRepository) CreateGroup(group *model.BOMChoiceGroup) error {
	return r.db.Create(group).Error
}

func (r *choiceRepository) DeleteGroup(id uint) error {
	// Delete options first, then group
	if err := r.db.Where("group_id = ?", id).Delete(&model.BOMChoiceOption{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&model.BOMChoiceGroup{}, id).Error
}

func (r *choiceRepository) AddOption(option *model.BOMChoiceOption) error {
	return r.db.Create(option).Error
}

func (r *choiceRepository) DeleteOption(id uint) error {
	return r.db.Delete(&model.BOMChoiceOption{}, id).Error
}
