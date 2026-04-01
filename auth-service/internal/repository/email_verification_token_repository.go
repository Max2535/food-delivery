package repository

import (
	"auth-service/internal/model"
	"time"

	"gorm.io/gorm"
)

type EmailVerificationTokenRepository interface {
	Create(token *model.EmailVerificationToken) error
	FindByTokenHash(hash string) (*model.EmailVerificationToken, error)
	DeleteByTokenHash(hash string) error
	DeleteByUserID(userID uint) error
	DeleteExpired() error
}

type emailVerificationTokenRepository struct {
	db *gorm.DB
}

func NewEmailVerificationTokenRepository(db *gorm.DB) EmailVerificationTokenRepository {
	return &emailVerificationTokenRepository{db: db}
}

func (r *emailVerificationTokenRepository) Create(token *model.EmailVerificationToken) error {
	return r.db.Create(token).Error
}

func (r *emailVerificationTokenRepository) FindByTokenHash(hash string) (*model.EmailVerificationToken, error) {
	var token model.EmailVerificationToken
	err := r.db.Where("token_hash = ?", hash).
		Preload("User").
		First(&token).Error
	return &token, err
}

func (r *emailVerificationTokenRepository) DeleteByTokenHash(hash string) error {
	return r.db.Where("token_hash = ?", hash).
		Delete(&model.EmailVerificationToken{}).Error
}

func (r *emailVerificationTokenRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).
		Delete(&model.EmailVerificationToken{}).Error
}

func (r *emailVerificationTokenRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).
		Delete(&model.EmailVerificationToken{}).Error
}
