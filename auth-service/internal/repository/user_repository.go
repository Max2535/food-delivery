package repository

import (
	"auth-service/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *model.User) error
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
	FindByIDs(ids []uint) ([]model.User, error)
	UpdatePassword(id uint, hashedPassword string) error
	UpdateGroupID(userIDs []uint, groupID uint) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Group").Preload("Group.Roles").Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Group").Preload("Group.Roles").Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Group").Preload("Group.Roles").First(&user, id).Error
	return &user, err
}

func (r *userRepository) FindByIDs(ids []uint) ([]model.User, error) {
	var users []model.User
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	return users, err
}

func (r *userRepository) UpdateGroupID(userIDs []uint, groupID uint) error {
	return r.db.Model(&model.User{}).Where("id IN ?", userIDs).Update("group_id", groupID).Error
}

func (r *userRepository) UpdatePassword(id uint, hashedPassword string) error {
	return r.db.Model(&model.User{}).
		Where("id = ?", id).
		Update("password", hashedPassword).Error
}
