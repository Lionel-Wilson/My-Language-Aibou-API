package main

import (
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/config"
	middlewares "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/middleware"
	sentencehandler "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/sentence"
	wordhandler "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/word"
	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/services/sentence"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/services/word"
	logger2 "github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	logger := logger2.New(cfg)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	store := cookie.NewStore([]byte(cfg.Secret))
	store.Options(sessions.Options{
		MaxAge:   12 * 60 * 60, // 12 hours
		HttpOnly: true,
		Secure:   true, // true in production
	})

	router.Use(sessions.Sessions("mysession", store))
	router.Use(middlewares.SecureHeaders())
	router.Use(middlewares.CorsMiddleware())

	openAiClient := openai.NewClient(cfg.OpenAIAPIKey, logger)

	wordService := word.NewWordService(logger, openAiClient)
	sentenceService := sentence.New(logger, openAiClient)

	wordHandler := wordhandler.NewWordHandler(logger, wordService)
	sentenceHandler := sentencehandler.NewSentenceHandler(logger, sentenceService)

	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/search/word", wordHandler.DefineWord)
		apiV1.POST("/search/synonyms", wordHandler.GetSynonyms)

		apiV1.POST("/search/sentence", sentenceHandler.ExplainSentence)
		apiV1.POST("/search/sentence/correction", sentenceHandler.CorrectSentence)
	}
	logger.Sugar().Infof("Server starting on port %s", cfg.Address)

	// router.RunTLS(addr, "./tls/cert.pem", "./tls/key.pem") TO-DO: Server over HTTPS when figure out how to get certificates
	err = router.Run(cfg.Address)
	if err != nil {
		logger.Fatal(err.Error())
	}
}
