package model

import "time"

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"index;not null"`
	TokenHash string    `gorm:"uniqueIndex;not null"` // SHA-256 hash ของ raw token
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	User      User `gorm:"foreignKey:UserID" json:"-"`
}

func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}
