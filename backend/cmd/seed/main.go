package main

import (
	"log"
	"os"
	"strings"

	"github.com/AiEmStylix/ecommerce/backend/internal/config"
	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"github.com/AiEmStylix/ecommerce/backend/internal/seeders"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found. Falling back to system environment variables.")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading variables: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DBSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("Cannot connect to Database:", err)
	}

	db.AutoMigrate(
		&models.University{},
		&models.User{},
		&models.Category{},
		&models.Location{},
		&models.Item{},
		&models.ItemImage{},
		&models.Offer{},
		&models.Appointment{},
		&models.AuctionConfig{},
		&models.Bid{},
		&models.ShippingService{},
		&models.Notification{},
	)

	force := strings.EqualFold(os.Getenv("SEED_FORCE"), "true")
	seeders.SeedData(db, seeders.Options{Force: force})
}
