package initializers

import (
	"authSystem/models"
	"log"
)

func SyncDatabase() error {
	if DB == nil {
		log.Fatal("Database connection not initialized")
	}

	err := DB.AutoMigrate(
		&models.User{},
		&models.Book{},
	)
	
	if err != nil {
		log.Fatal("Failed to sync database: ", err)
		return err
	}

	log.Println("Database synced successfully")
	return nil
}