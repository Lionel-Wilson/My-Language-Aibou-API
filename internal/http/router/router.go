package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/auth"
	sentencehandler "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/sentence"
	subscriptions2 "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/subscriptions"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/webhook"
	wordhandler "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/word"
	auth2 "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/auth"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/sentence"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/subscriptions"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/word"
	commonMiddleware "github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/middleware"
	"github.com/Lionel-Wilson/My-Language-Aibou-API/pkg/commonlibrary/render"
)

func New(
	logger *zap.Logger,
	wordService word.Service,
	sentenceService sentence.Service,
	userService auth2.UserService,
	subscriptionService subscriptions.SubscriptionService,
	jwtSecret []byte,
	stripeWebhookSecret string,
) http.Handler {
	// Create a new Chi router.
	router := chi.NewRouter()

	// Add middleware.
	router.Use(middleware.Logger)    // logs every request
	router.Use(middleware.Recoverer) // recovers from panics

	// Define the /alive endpoint.
	registerAliveEndpoint(router)

	authHandler := auth.NewAuthHandler(logger, userService, subscriptionService)
	wordHandler := wordhandler.NewWordHandler(logger, wordService)
	sentenceHandler := sentencehandler.NewSentenceHandler(logger, sentenceService)
	subscriptionsHandler := subscriptions2.NewSubscriptionsHandler(logger, subscriptionService, userService)
	webhookHandler := webhook.NewWebhookHandler(logger, stripeWebhookSecret, subscriptionService)

	router.Route(
		"/api/v1", func(r chi.Router) {
			r.Route(
				"/search", func(r chi.Router) {
					r.Post("/word", wordHandler.DefineWord())
					r.Post("/synonyms", wordHandler.GetSynonyms())
					r.Post("/sentence", sentenceHandler.ExplainSentence())
					r.Post("/sentence/correction", sentenceHandler.CorrectSentence())
				},
			)
		},
	)

	router.Route(
		"/api/v2", func(r chi.Router) {
			r.Route(
				"/word", func(r chi.Router) {
					r.Post("/definition", wordHandler.DefineWord())
					r.Post("/synonyms", wordHandler.GetSynonyms())
					r.Post("/history", wordHandler.GetHistory())
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

	router.Route("/api/v3", func(r chi.Router) {
		r.Post("/webhooks/stripe", webhookHandler.HandleStripeWebhook())

		r.Route(
			"/auth", func(r chi.Router) {
				r.Post("/register", authHandler.Register())
				r.Post("/login", authHandler.Login())
			},
		)

		// Protected endpoints: wrap these with auth middleware.
		r.Group(func(r chi.Router) {
			r.Use(commonMiddleware.AuthMiddlewareString(jwtSecret))
			r.Route(
				"/user", func(r chi.Router) {
					r.Post("/update-details", authHandler.UpdateDetails())
					r.Delete("/", authHandler.Delete())
				})

			r.Route(
				"/word", func(r chi.Router) {
					r.Post("/definition", wordHandler.DefineWord())
					r.Post("/synonyms", wordHandler.GetSynonyms())
					r.Post("/history", wordHandler.GetHistory())
				},
			)
			r.Route(
				"/sentence", func(r chi.Router) {
					r.Post("/explanation", sentenceHandler.ExplainSentence())
					r.Post("/correction", sentenceHandler.CorrectSentence())
				},
			)

			r.Route(
				"/subscription", func(r chi.Router) {
					r.Post("/subscribe", subscriptionsHandler.Subscribe())
					r.Post("/cancel", subscriptionsHandler.Cancel())
					r.Get("/status", subscriptionsHandler.Status())
					r.Post("/checkout", subscriptionsHandler.CreateCheckoutSession())
				},
			)
		})
	})

	return router
}

func registerAliveEndpoint(router *chi.Mux) {
	router.Get("/alive", func(w http.ResponseWriter, r *http.Request) {
		// Return a simple status message.
		render.Json(w, http.StatusOK, "API is alive!")
	})
}
