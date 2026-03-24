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
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"-"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	RoleID   uint   `gorm:"not null" json:"role_id"`
	Role     Role   `gorm:"foreignKey:RoleID" json:"role"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
