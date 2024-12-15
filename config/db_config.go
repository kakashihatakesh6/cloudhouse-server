package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadConfig() {
	err := godotenv.Load()

	if err != nil {
		log.Println("Error loading .env file")
	}

	requiredVars := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"}

	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Fatalf("Missing required enviornment variables %s", v)
		}
	}

}