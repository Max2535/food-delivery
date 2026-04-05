package repository

import (
	"auth-service/internal/model"
	"gorm.io/gorm"
	"time"
)

type UserRepository interface {
	Create(user *model.User) error
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
	FindByIDs(ids []uint) ([]model.User, error)
	ListAll() ([]model.User, error)
	UpdatePassword(id uint, hashedPassword string) error
	UpdateGroupID(userIDs []uint, groupID uint) error
	UpdateIsVerified(id uint, isVerified bool) error
	Delete(id uint) error
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
	err := r.db.Preload("Group").Preload("Group.Roles").Preload("Group.Roles.Permissions").Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Group").Preload("Group.Roles").Preload("Group.Roles.Permissions").Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Group").Preload("Group.Roles").Preload("Group.Roles.Permissions").First(&user, id).Error
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

func (r *userRepository) UpdateIsVerified(id uint, isVerified bool) error {
	updates := map[string]any{"is_verified": isVerified}
	if isVerified {
		now := time.Now()
		updates["verified_at"] = now
	} else {
		updates["verified_at"] = nil
	}
	return r.db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *userRepository) ListAll() ([]model.User, error) {
	var users []model.User
	err := r.db.Preload("Group").Find(&users).Error
	return users, err
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}
