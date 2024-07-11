package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

// CorsMiddleware sets up the CORS (Cross-Origin Resource Sharing) middleware.
// It allows requests from specified origins, methods, and headers, and supports credentials.
// This middleware should be added to the Gin router to handle CORS for incoming requests.
func CorsMiddleware() gin.HandlerFunc {
	// Define the CORS configuration options
	corsConfig := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	return func(c *gin.Context) {
		// Apply the CORS configuration to the current request
		corsConfig.HandlerFunc(c.Writer, c.Request)
		// Proceed to the next middleware or handler
		c.Next()
	}
}
