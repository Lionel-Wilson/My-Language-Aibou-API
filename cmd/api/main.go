package main

import (
	"fmt"
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

	openAiClient := openai.NewClient(cfg.OpenAIAPIKey, logger)

	wordService := word.NewWordService(logger, openAiClient)
	sentenceService := sentence.NewSentenceService(logger, openAiClient)

	mux := router.New(logger, wordService, sentenceService)

	logger.Sugar().Infof("Server starting on port %s", cfg.Port)

	addr := fmt.Sprintf(":%s", cfg.Port)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
