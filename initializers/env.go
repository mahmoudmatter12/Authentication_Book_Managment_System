package initializers

import (
	"github.com/joho/godotenv"
	"os"
	"fmt"
)

func LoadEnvVars() error {
    if err := godotenv.Load(); err != nil {
        return fmt.Errorf("failed to load .env file: %v\nCurrent working directory: %s\nIs .env file present? %v",
            err,
            os.Getenv("PWD"),
            fileExists(".env"))
    }
    return nil
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}
