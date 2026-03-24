package repository

import (
	"auth-service/internal/model"

	"gorm.io/gorm"
)

type GroupRepository interface {
	FindByName(name string) (*model.Group, error)
	ListAll() ([]*model.Group, error)
	ListRoles() ([]*model.Role, error)
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

func (r *groupRepository) ListAll() ([]*model.Group, error) {
	var groups []*model.Group
	err := r.db.Preload("Roles").Find(&groups).Error
	return groups, err
}

func (r *groupRepository) ListRoles() ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.Find(&roles).Error
	return roles, err
}
