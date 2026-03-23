package model

import "time"

type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"index;not null"`
	TokenHash string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	User      User `gorm:"foreignKey:UserID" json:"-"`
}

func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
