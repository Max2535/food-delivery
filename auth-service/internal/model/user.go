package model

import "gorm.io/gorm"

const (
	RoleUser     = "user"
	RoleAdmin    = "admin"
	RoleRider    = "rider"
	RoleMerchant = "merchant"
	RoleCustomer = "customer"
)

type User struct {
	gorm.Model
	Username   string `gorm:"uniqueIndex;not null" json:"username"`
	Password   string `gorm:"not null" json:"-"`
	Email      string `gorm:"uniqueIndex;not null" json:"email"`
	GroupID    uint   `gorm:"not null" json:"group_id"`
	Group      Group  `gorm:"foreignKey:GroupID" json:"group"`
	IsVerified bool   `gorm:"default:true" json:"is_verified"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
