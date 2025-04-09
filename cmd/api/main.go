package main

import (
	"fmt"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth/storage"
	commonDb "github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/db"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // <-- Add this line to register the Postgres driver
	"log"
	"net/http"

	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/config"
	router "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/http/router"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/sentence"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/word"
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

	userRepository := storage.NewUserRepository(db)
	userService := auth.NewUserService(logger, userRepository, cfg.JwtSecret)

	mux := router.New(
		logger,
		wordService,
		sentenceService,
		userService,
		cfg.JwtSecret,
	)

	logger.Sugar().Infof("Server starting on port %s", cfg.Port)

	addr := fmt.Sprintf(":%s", cfg.Port)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
