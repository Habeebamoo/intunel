package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                     string
	Env                      string
	RedisUrl                 string
	PostgresUrl              string
	ResendApiKey             string
}

func Load() *Config {
	// load .env file (ignore error in prod)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	return &Config{
		Port:                     port,
		Env:                      env,
		RedisUrl: 					      redisURL,
		PostgresUrl:              os.Getenv("POSTGRES_URL"),
		ResendApiKey: 					 	os.Getenv("RESEND_API_KEY"),
	}
}