package repository

import (
	"auth-service/internal/model"
	"time"

	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *model.RefreshToken) error
	FindByTokenHash(hash string) (*model.RefreshToken, error)
	DeleteByTokenHash(hash string) error
	DeleteByUserID(userID uint) error
	DeleteExpired() error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) FindByTokenHash(hash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.db.Where("token_hash = ?", hash).
		Preload("User").
		Preload("User.Group").
		Preload("User.Group.Roles").
		First(&token).Error
	return &token, err
}

func (r *refreshTokenRepository) DeleteByTokenHash(hash string) error {
	return r.db.Where("token_hash = ?", hash).
		Delete(&model.RefreshToken{}).Error
}

func (r *refreshTokenRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).
		Delete(&model.RefreshToken{}).Error
}

func (r *refreshTokenRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).
		Delete(&model.RefreshToken{}).Error
}
