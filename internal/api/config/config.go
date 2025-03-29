package config

import (
	"fmt"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds the application settings.
type Config struct {
	OpenAIAPIKey string `mapstructure:"OPENAI_API_KEY" validate:"required"`
	Secret       string `mapstructure:"SECRET" validate:"required"`
	Port         string `mapstructure:"PORT" validate:"required"`
	Env          string `mapstructure:"ENV" validate:"required"`
}

// LoadConfig loads configuration from the OS environment and, if not in production,
// from a .env file at the root of the repository.
func LoadConfig() (*Config, error) {
	// Check if running in production.
	// When ENV is "prod", we assume all necessary environment variables are set.
	// Otherwise, load variables from the .env file.
	if os.Getenv("ENV") != "prod" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: no .env file found, relying on OS environment variables: %v", err)
		}
	}

	// Use Viper to read environment variables.
	viper.AutomaticEnv()

	// Set a default value for ENV if it hasn't been set.
	if viper.GetString("ENV") == "" {
		viper.Set("ENV", "dev")
	}

	// Create a Config instance with values from environment variables.
	cfg := Config{
		OpenAIAPIKey: viper.GetString("OPENAI_API_KEY"),
		Secret:       viper.GetString("SECRET"),
		Env:          viper.GetString("ENV"),
		Port:         viper.GetString("PORT"),
	}

	// Validate the config.
	if err := validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
