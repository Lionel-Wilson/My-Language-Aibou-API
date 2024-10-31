package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenAi struct {
		Key string
	}
	Secret  string
	Address string
}

func New() *Config {
	//Load environment variables. Uncomment when running locally and not in container TO-DO: put this in an if env="local" statement in the config NEW function
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var cfg Config

	addr := os.Getenv("DEV_ADDRESS")
	secret := os.Getenv("SECRET")
	openAiKey := os.Getenv("OPENAI_API_KEY")

	cfg.Address = addr
	cfg.Secret = secret
	cfg.OpenAi.Key = openAiKey

	return &cfg
}
