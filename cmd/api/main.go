package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // <-- Add this line to register the Postgres driver

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth"
	userStorage "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth/storage"
	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/config"
	router "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/http/router"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/sentence"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/subscriptions"
	subscriptionStorage "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/subscriptions/storage"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/word"
	commonDb "github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/db"
	commonlogger "github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	logger := commonlogger.New(cfg)

	// Connect to the PostgreSQL database using sqlx.
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Sugar().Fatalf("failed to connect to database: %v", err)
	}

	if err := commonDb.RunMigrations(db.DB); err != nil {
		logger.Sugar().Fatalf("failed to run migrations: %v", err)
	}

	openAiClient := openai.NewClient(cfg.OpenAIAPIKey, logger)

	wordService := word.NewWordService(logger, openAiClient)
	sentenceService := sentence.NewSentenceService(logger, openAiClient)

	userRepository := userStorage.NewUserRepository(db)
	userService := auth.NewUserService(logger, userRepository, cfg.JwtSecret, cfg.StripeSecretKey)

	subscriptionRepository := subscriptionStorage.NewSubscriptionsRepository(db)
	subscriptionService := subscriptions.NewSubscriptionService(logger, cfg.StripeSecretKey, subscriptionRepository)

	mux := router.New(
		logger,
		wordService,
		sentenceService,
		userService,
		subscriptionService,
		cfg.JwtSecret,
	)

	logger.Sugar().Infof("Server starting on port %s", cfg.Port)

	addr := fmt.Sprintf(":%s", cfg.Port)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
