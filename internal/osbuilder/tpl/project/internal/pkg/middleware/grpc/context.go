package grpc

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"{{.D.ModuleName}}/internal/pkg/contextx"
)

// Context creates a unary server interceptor that injects needed keys into the context.
func Context() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		spanCtx := trace.SpanFromContext(ctx).SpanContext()

		// Only inject traceID if span context is valid
		if spanCtx.IsValid() {
			traceID := spanCtx.TraceID().String()
			ctx = contextx.WithTraceID(ctx, traceID)
		}

		return handler(ctx, req)
	}
}
