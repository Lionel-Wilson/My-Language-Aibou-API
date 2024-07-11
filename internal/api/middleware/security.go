package middlewares

import "github.com/gin-gonic/gin"

// SecureHeaders sets security-related HTTP headers to enhance the security of the application.
// It adds the following headers:
// - X-XSS-Protection: Enables cross-site scripting (XSS) protection.
// - X-Frame-Options: Prevents the page from being displayed in an iframe (clickjacking protection).
// This middleware should be added to the Gin router to apply these headers to all responses.
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set the X-XSS-Protection header to enable XSS protection
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		// Set the X-Frame-Options header to deny framing of the page
		c.Writer.Header().Set("X-Frame-Options", "deny")
		// Proceed to the next middleware or handler
		c.Next()
	}
}
