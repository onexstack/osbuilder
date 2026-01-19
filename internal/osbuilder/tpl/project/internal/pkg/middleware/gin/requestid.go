package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"{{.D.ModuleName}}/internal/pkg/contextx"
	"{{.D.ModuleName}}/internal/pkg/known"
)

// RequestIDMiddleware is a Gin middleware that ensures every HTTP request
// has a unique request ID. It extracts the `x-request-id` from the incoming
// request headers, or generates a new UUID if one is not present.
// This request ID is then injected into the request's `context.Context` for
// downstream access and also added to the HTTP response headers.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Attempt to retrieve the request ID from the incoming `x-request-id` header.
		requestID := c.Request.Header.Get(string(known.XRequestID))

		// If no request ID is provided in the header, generate a new UUID.
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store the request ID in the request's context.Context using `contextx`.
		// This makes the request ID accessible to subsequent handlers and business logic.
		ctx := contextx.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		// Set the `x-request-id` header in the HTTP response, ensuring the client
		// receives the request ID for tracing and correlation.
		c.Writer.Header().Set(string(known.XRequestID), requestID)

		// Proceed to the next middleware or handler in the Gin chain.
		c.Next()
	}
}
