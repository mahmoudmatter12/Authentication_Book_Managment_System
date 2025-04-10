package initializers

import (
	"log"
	"os"
	"sync"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB   *gorm.DB
	dbOnce sync.Once
)

func ConnectToDB() error {
	var err error
	var initErr error

	dbOnce.Do(func() {
		dsn := os.Getenv("DB_DSN")
		if dsn == "" {
			initErr = fmt.Errorf("DB_DSN environment variable not set")
			log.Fatal(initErr)
			return
		}

		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			initErr = fmt.Errorf("Failed to connect to database: %v", err)
			log.Fatal(initErr)
			return
		}

		log.Println("Connected to database successfully")
	})

	return initErr
}