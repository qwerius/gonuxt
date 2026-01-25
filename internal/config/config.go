package config
import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

func Load(){
	if err := godotenv.Load(".env"); err != nil {
		log.Println(".env not loaded (using OS env)")
	}
}

func Get(key string) string {
	return os.Getenv(key)
}