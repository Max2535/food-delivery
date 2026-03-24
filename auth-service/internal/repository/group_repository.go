package repository

import (
	"auth-service/internal/model"
	"gorm.io/gorm"
)

type GroupRepository interface {
	FindByName(name string) (*model.Group, error)
}

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) FindByName(name string) (*model.Group, error) {
	var group model.Group
	err := r.db.Preload("Roles").Where("name = ?", name).First(&group).Error
	return &group, err
}
