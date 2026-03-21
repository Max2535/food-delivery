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
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Mock Data
	users := []model.User{
		{
			Username: "admin",
			Password: "password",
			Email:    "admin@example.com",
			Role:     model.RoleAdmin,
		},
		{
			Username: "rider_01",
			Password: "securepassword123",
			Email:    "rider01@example.com",
			Role:     model.RoleRider,
		},
		{
			Username: "user",
			Password: "password",
			Email:    "user@example.com",
			Role:     model.RoleUser,
		},
		{
			Username: "validuser",
			Password: "validpassword",
			Email:    "validuser@example.com",
			Role:     model.RoleUser,
		},
	}

	log.Println("Seeding data...")

	for _, user := range users {
		// Hash password before saving
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash password for user %s: %v", user.Username, err)
			continue
		}
		user.Password = string(hashedPassword)

		// Check if user already exists
		var existingUser model.User
		if err := db.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
			// Update existing user
			user.ID = existingUser.ID
			if err := db.Save(&user).Error; err != nil {
				log.Printf("Failed to update user %s: %v", user.Username, err)
			} else {
				log.Printf("Updated user %s", user.Username)
			}
		} else {
			// Create new user
			if err := db.Create(&user).Error; err != nil {
				log.Printf("Failed to seed user %s: %v", user.Username, err)
			} else {
				log.Printf("Seeded user %s", user.Username)
			}
		}
	}

	log.Println("Seeding complete!")
}
