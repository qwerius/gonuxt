package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Load() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println(".env not loaded (using OS env)")
	}
}

func Get(key string) string {
	val := os.Getenv(key)
	if val == "" {
		if key == "ENV" {
			// fallback default development
			return "development"
		}
		// bisa tambahkan fallback lain untuk key lain jika perlu
	}
	return val
}
