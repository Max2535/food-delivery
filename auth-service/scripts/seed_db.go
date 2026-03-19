package scripts

import (
	"log"
	"os"

	"auth-service/internal/model"

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
	tickets := []model.User{
		{
			Username: "admin",
			Password: "password",
		},
		{
			Username: "user",
			Password: "password",
		},
	}

	log.Println("Seeding data...")

	for _, ticket := range tickets {
		if err := db.Create(&ticket).Error; err != nil {
			log.Printf("Failed to seed user %d: %v", ticket.ID, err)
		} else {
			log.Printf("Seeded user ID %d for user %d", ticket.ID, ticket.ID)
		}
	}

	log.Println("Seeding complete!")
}
