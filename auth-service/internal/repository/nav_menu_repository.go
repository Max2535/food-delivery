package repository

import (
	"auth-service/internal/model"

	"gorm.io/gorm"
)

type NavMenuRepository interface {
	ListAll() ([]model.NavGroup, error)
	FindGroupByLabel(label string) (*model.NavGroup, error)
	CreateGroup(group *model.NavGroup) error
	CreateItem(item *model.NavItem) error
}

type navMenuRepository struct {
	db *gorm.DB
}

func NewNavMenuRepository(db *gorm.DB) NavMenuRepository {
	return &navMenuRepository{db: db}
}

func (r *navMenuRepository) ListAll() ([]model.NavGroup, error) {
	var groups []model.NavGroup
	err := r.db.
		Preload("Roles").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Items.Roles").
		Order("sort_order ASC").
		Find(&groups).Error
	return groups, err
}

func (r *navMenuRepository) FindGroupByLabel(label string) (*model.NavGroup, error) {
	var group model.NavGroup
	err := r.db.Where("label = ?", label).First(&group).Error
	return &group, err
}

func (r *navMenuRepository) CreateGroup(group *model.NavGroup) error {
	return r.db.Create(group).Error
}

func (r *navMenuRepository) CreateItem(item *model.NavItem) error {
	return r.db.Create(item).Error
}
