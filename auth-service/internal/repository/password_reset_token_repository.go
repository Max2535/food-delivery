package repository

import (
	"auth-service/internal/model"
	"time"

	"gorm.io/gorm"
)

type PasswordResetTokenRepository interface {
	Create(token *model.PasswordResetToken) error
	FindByTokenHash(hash string) (*model.PasswordResetToken, error)
	DeleteByTokenHash(hash string) error
	DeleteByUserID(userID uint) error
	DeleteExpired() error
}

type passwordResetTokenRepository struct {
	db *gorm.DB
}

func NewPasswordResetTokenRepository(db *gorm.DB) PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(token *model.PasswordResetToken) error {
	return r.db.Create(token).Error
}

func (r *passwordResetTokenRepository) FindByTokenHash(hash string) (*model.PasswordResetToken, error) {
	var token model.PasswordResetToken
	err := r.db.Where("token_hash = ?", hash).
		Preload("User").
		First(&token).Error
	return &token, err
}

func (r *passwordResetTokenRepository) DeleteByTokenHash(hash string) error {
	return r.db.Where("token_hash = ?", hash).
		Delete(&model.PasswordResetToken{}).Error
}

func (r *passwordResetTokenRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).
		Delete(&model.PasswordResetToken{}).Error
}

func (r *passwordResetTokenRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).
		Delete(&model.PasswordResetToken{}).Error
}
