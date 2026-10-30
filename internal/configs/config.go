package configs

import (
	"fmt"
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

var Envs []string

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system env")
	}

	port := "PORT"
	env := "ENV"
	postgresUrl := "POSTGRES_URL"
	redisURL := "REDIS_URL"
	resendApiKey := "RESEND_API_KEY"
	senderDomain := "SENDER_DOMAIN"

	Envs =  append(Envs, 
		port,
		env,
		postgresUrl,
		redisURL,
		resendApiKey,
		senderDomain,
	)
	for _, env := range Envs {
		if err := Check(env); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	return &Config{
		Port:           os.Getenv(port),
		Env:            os.Getenv(env),
		RedisUrl: 		  os.Getenv(redisURL),
		PostgresUrl:    os.Getenv(postgresUrl),
		ResendApiKey: 	os.Getenv(resendApiKey),
	}
}

func Check(envName string) error {
	env := os.Getenv(envName)
	if env == "" {
		return fmt.Errorf("%v is missing", envName)
	}

	return nil
}