package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/coocood/freecache"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // <-- Add this line to register the Postgres driver

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth"
	authStorage "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth/storage"
	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/config"
	router "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/http/router"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/paymenttransactions"
	ptStorage "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/paymenttransactions/storage"
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

	// In bytes, where 1024 * 1024 represents a single Megabyte, and 100 * 1024*1024 represents 100 Megabytes.
	cacheSize := 100 * 1024 * 1024
	cache := freecache.NewCache(cacheSize)

	openAiClient := openai.NewClient(cfg.OpenAIAPIKey, logger)

	wordService := word.NewWordService(logger, openAiClient, cache) // todo: make a db to store words and sentences rather than a cache
	sentenceService := sentence.NewSentenceService(logger, openAiClient, cache)

	userRepository := authStorage.NewUserRepository(db)
	userService := auth.NewUserService(logger, userRepository, cfg.JwtSecret, cfg.StripeSecretKey)

	paymentTransactionsRepository := ptStorage.NewPaymentTransactionRepository(db)
	paymentTransactionService := paymenttransactions.NewPaymentTransactionService(logger, paymentTransactionsRepository)

	subscriptionRepository := subscriptionStorage.NewSubscriptionsRepository(db)
	subscriptionService := subscriptions.NewSubscriptionService(
		logger,
		cfg.StripeSecretKey,
		subscriptionRepository,
		paymentTransactionService,
		userService,
		cfg.StripePaidPriceId,
		cfg.CheckoutSuccessURL,
		cfg.CheckoutCancelURL,
	)

	mux := router.New(
		logger,
		wordService,
		sentenceService,
		userService,
		subscriptionService,
		cfg.JwtSecret,
		cfg.StripeWebhookSecret,
	)

	logger.Sugar().Infof("Server starting on port %s", cfg.Port)

	addr := fmt.Sprintf(":%s", cfg.Port)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
