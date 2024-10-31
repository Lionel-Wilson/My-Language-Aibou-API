package main

import (
	log "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/log"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/config"
	handler "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/language_tools"
	middlewares "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/middleware"
	languagetools "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/language_tools"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.New()
	logger := log.New()

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

	languageToolsService := languagetools.New(cfg, logger)
	languageToolsHandler := handler.NewHandler(logger, languageToolsService)

	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/search/word", languageToolsHandler.DefineWord)
		apiV1.POST("/search/sentence", languageToolsHandler.ExplainSentence)
		apiV1.POST("/search/synonyms", languageToolsHandler.GetSynonyms)
	}
	logger.InfoLog.Printf("Starting server on %s", cfg.Address)

	//router.RunTLS(addr, "./tls/cert.pem", "./tls/key.pem") TO-DO: Server over HTTPS when figure out how to get certificates
	err := router.Run(cfg.Address)
	if err != nil {
		logger.ErrorLog.Fatal(err)
	}
}
