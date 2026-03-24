package main

import (
	"log"
	"os"

	"auth-service/internal/model"
	"golang.org/x/crypto/bcrypt"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Database Connection
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not set in .env")
	}

	log.Printf("Connecting to database at %s...", dbURL)
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection successful.")

	// AutoMigrate the schema
	log.Println("Migrating database...")
	if err := db.AutoMigrate(&model.Role{}, &model.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Ensure roles exist
	roleMap := make(map[string]uint)
	for _, name := range []string{model.RoleAdmin, model.RoleRider, model.RoleCustomer, model.RoleUser, model.RoleMerchant} {
		var role model.Role
		if err := db.Where("name = ?", name).First(&role).Error; err != nil {
			role = model.Role{Name: name}
			db.Create(&role)
		}
		roleMap[name] = role.ID
	}

	// Mock Data
	type seedUser struct {
		Username string
		Password string
		Email    string
		RoleName string
	}
	users := []seedUser{
		{"admin", "password", "admin@example.com", model.RoleAdmin},
		{"rider_01", "securepassword123", "rider01@example.com", model.RoleRider},
		{"user", "password", "user@example.com", model.RoleUser},
		{"validuser", "validpassword", "validuser@example.com", model.RoleUser},
	}

	log.Println("Seeding data...")

	for _, u := range users {
		// Hash password before saving
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash password for user %s: %v", u.Username, err)
			continue
		}

		user := model.User{
			Username: u.Username,
			Password: string(hashedPassword),
			Email:    u.Email,
			RoleID:   roleMap[u.RoleName],
		}

		// Check if user already exists
		var existingUser model.User
		if err := db.Where("username = ?", u.Username).First(&existingUser).Error; err == nil {
			// Update existing user
			user.ID = existingUser.ID
			if err := db.Save(&user).Error; err != nil {
				log.Printf("Failed to update user %s: %v", u.Username, err)
			} else {
				log.Printf("Updated user %s", u.Username)
			}
		} else {
			// Create new user
			if err := db.Create(&user).Error; err != nil {
				log.Printf("Failed to seed user %s: %v", u.Username, err)
			} else {
				log.Printf("Seeded user %s", u.Username)
			}
		}
	}

	log.Println("Seeding complete!")
}
