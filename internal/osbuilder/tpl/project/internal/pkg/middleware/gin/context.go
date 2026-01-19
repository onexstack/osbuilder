package gin

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"{{.D.ModuleName}}/internal/pkg/contextx"
)

// Context is a middleware that injects common prefix fields to gin.Context.
func Context() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the TraceID from the current OpenTelemetry span in the request context.
		// If no span is present, a zero-value TraceID will be returned, which is still a valid string.
		traceID := trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()

		// Create a new context with the extracted TraceID.
		// This makes the TraceID available via contextx.TraceID(ctx).
		ctx := contextx.WithTraceID(c.Request.Context(), traceID)

		// Update the Gin request's context with the new context containing the TraceID.
		c.Request = c.Request.WithContext(ctx)

		// Proceed to the next middleware or handler in the Gin chain.
		c.Next()
	}
}
