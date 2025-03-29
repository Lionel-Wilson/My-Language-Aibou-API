package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	sentencehandler "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/sentence"
	wordhandler "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/word"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/sentence"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/word"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/render"
)

func New(
	logger *zap.Logger,
	wordService word.Service,
	sentencService sentence.Service,
) http.Handler {
	// Create a new Chi router.
	router := chi.NewRouter()

	// Add middleware.
	router.Use(middleware.Logger)    // logs every request
	router.Use(middleware.Recoverer) // recovers from panics

	// Define the /alive endpoint.
	registerAliveEndpoint(router)
	router.Route(
		"/api/v1", func(r chi.Router) {
			wordHandler := wordhandler.NewWordHandler(logger, wordService)
			sentenceHandler := sentencehandler.NewSentenceHandler(logger, sentencService)

			r.Route(
				"/word", func(r chi.Router) {
					r.Post("/definition", wordHandler.DefineWord())
					r.Post("/synonyms", wordHandler.GetSynonyms())
				},
			)
			r.Route(
				"/sentence", func(r chi.Router) {
					r.Post("/explanation", sentenceHandler.ExplainSentence())
					r.Post("/correction", sentenceHandler.CorrectSentence())
				},
			)
		},
	)

	return router
}

func registerAliveEndpoint(router *chi.Mux) {
	router.Get("/alive", func(w http.ResponseWriter, r *http.Request) {
		// Return a simple status message.
		render.Json(w, http.StatusOK, "API is alive!")
	})
}
