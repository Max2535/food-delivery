package repository

import (
	"auth-service/internal/model"

	"gorm.io/gorm"
)

type GroupRepository interface {
	FindByName(name string) (*model.Group, error)
	ListAll() ([]*model.Group, error)
	ListRoles() ([]*model.Role, error)
	Create(group *model.Group) error
	FindByID(id uint) (*model.Group, error)
	Update(group *model.Group) error
	FindRolesByIDs(ids []uint) ([]model.Role, error)
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
	err := r.db.Preload("Roles").Preload("Users").Find(&groups).Error
	return groups, err
}

func (r *groupRepository) ListRoles() ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.Find(&roles).Error
	return roles, err
}

func (r *groupRepository) Create(group *model.Group) error {
	return r.db.Create(group).Error
}

func (r *groupRepository) FindByID(id uint) (*model.Group, error) {
	var group model.Group
	err := r.db.Preload("Roles").Preload("Users").First(&group, id).Error
	return &group, err
}

func (r *groupRepository) Update(group *model.Group) error {
	if err := r.db.Model(group).Association("Roles").Replace(group.Roles); err != nil {
		return err
	}
	return r.db.Save(group).Error
}

func (r *groupRepository) FindRolesByIDs(ids []uint) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Where("id IN ?", ids).Find(&roles).Error
	return roles, err
}
