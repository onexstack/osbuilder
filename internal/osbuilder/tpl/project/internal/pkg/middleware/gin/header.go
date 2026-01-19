package gin

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// NoCache is a Gin middleware that prevents clients from caching HTTP responses.
// It sets standard Cache-Control, Expires, and Last-Modified headers to
// ensure the client always fetches the fresh content.
func NoCache(c *gin.Context) {
	c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
	c.Header("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
	c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	c.Next()
}

// CorsConfig holds configuration options for the CORS middleware.
type CorsConfig struct {
	// AllowedOrigins specifies a list of origins that may access the resource.
	// If set to ["*"], all origins are allowed.
	AllowedOrigins []string
	// AllowedMethods specifies the methods allowed when accessing the resource.
	AllowedMethods []string
	// AllowedHeaders specifies the headers that can be used when making the actual request.
	AllowedHeaders []string
	// AllowCredentials indicates whether or not the response to the request
	// can be exposed when the credentials flag is true.
	AllowCredentials bool
	// MaxAge indicates how long the results of a preflight request can be cached.
	MaxAge time.Duration
}

// DefaultCorsConfig provides a sensible default configuration for CORS.
func DefaultCorsConfig() CorsConfig {
	return CorsConfig{
		AllowedOrigins:   []string{"*"}, // Allow all origins by default
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// Cors is a Gin middleware that handles Cross-Origin Resource Sharing (CORS) requests.
// It configures the necessary HTTP headers based on the provided CorsConfig.
func Cors(config CorsConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		var allowedOrigin string

		// Check if the origin is allowed
		if len(config.AllowedOrigins) == 0 || (len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*") {
			allowedOrigin = "*"
		} else {
			for _, o := range config.AllowedOrigins {
				if o == origin {
					allowedOrigin = origin
					break
				}
			}
		}

		// Set response headers for all requests
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		c.Header("Access-Control-Max-Age", fmt.Sprintf("%.0f", config.MaxAge.Seconds()))

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight requests (OPTIONS method)
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent) // Use 204 No Content for preflight success
			return
		}

		c.Next() // Continue processing the actual request
	}
}

// Secure is a Gin middleware that adds security-related HTTP headers to responses.
// These headers help mitigate common web vulnerabilities.
func Secure(c *gin.Context) {
	c.Header("X-Frame-Options", "DENY")           // Prevents clickjacking by forbidding embedding in iframes
	c.Header("X-Content-Type-Options", "nosniff") // Prevents browsers from MIME-sniffing a response away from the declared Content-Type
	c.Header("X-XSS-Protection", "1; mode=block") // Enables the XSS filter in browsers

	// Only apply HSTS over TLS connections
	if c.Request.TLS != nil {
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	}
	c.Next()
}
