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
	OpenAIAPIKey        string `mapstructure:"OPENAI_API_KEY" validate:"required"`
	Secret              string `mapstructure:"SECRET" validate:"required"`
	Port                string `mapstructure:"PORT" validate:"required"`
	Env                 string `mapstructure:"ENV" validate:"required"`
	DatabaseURL         string `mapstructure:"DATABASE_URL" yaml:"database_url" validate:"required"`
	LogLevel            string `mapstructure:"LOG_LEVEL" yaml:"log_level"`
	JwtSecret           []byte `mapstructure:"JWT_SECRET" yaml:"jwt_secret" validate:"required"`
	StripeSecretKey     string `mapstructure:"STRIPE_SECRET_KEY" yaml:"stripe_secret_key" validate:"required"`
	StripeWebhookSecret string `mapstructure:"STRIPE_WEBHOOK_SECRET" yaml:"webhook_secret" validate:"required"`
	StripePaidPriceId   string `mapstructure:"STRIPE_PAID_PRICE_ID" yaml:"stripe_paid_price_id" validate:"required"`
	CheckoutSuccessURL  string `mapstructure:"CHECKOUT_SUCCESS_URL" yaml:"checkout_success_url" validate:"required"`
	CheckoutCancelURL   string `mapstructure:"CHECKOUT_CANCEL_URL" yaml:"checkout_cancel_url" validate:"required"`
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
		OpenAIAPIKey:        viper.GetString("OPENAI_API_KEY"),
		Secret:              viper.GetString("SECRET"),
		Env:                 viper.GetString("ENV"),
		Port:                viper.GetString("PORT"),
		DatabaseURL:         viper.GetString("DATABASE_URL"),
		LogLevel:            viper.GetString("LOG_LEVEL"),
		StripeSecretKey:     viper.GetString("STRIPE_SECRET_KEY"),
		JwtSecret:           []byte(viper.GetString("JWT_SECRET")),
		StripeWebhookSecret: viper.GetString("STRIPE_WEBHOOK_SECRET"),
		StripePaidPriceId:   viper.GetString("STRIPE_PAID_PRICE_ID"),
		CheckoutSuccessURL:  viper.GetString("CHECKOUT_SUCCESS_URL"),
		CheckoutCancelURL:   viper.GetString("CHECKOUT_CANCEL_URL"),
	}

	// Validate the config.
	if err := validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
