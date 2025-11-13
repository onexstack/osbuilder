package contextx

import (
	"context"
	"log/slog"
)

// Define keys for the context.
type (
	// usernameKey defines the context key for the username.
	usernameKey struct{}
	// userIDKey defines the context key for the user ID.
	userIDKey struct{}
	// accessTokenKey defines the context key for the access token.
	accessTokenKey struct{}
	// requestIDKey defines the context key for the request ID.
	requestIDKey struct{}
    // traceIDKey is the key for storing trace ID in context
    traceIDKey struct{}
    // loggerKey is the key for storing logger in context
    loggerKey struct{}
)

// WithUserID stores the user ID into the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// UserID retrieves the user ID from the context.
func UserID(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey{}).(string)
	return userID
}

// WithUsername stores the username into the context.
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey{}, username)
}

// Username retrieves the username from the context.
func Username(ctx context.Context) string {
	username, _ := ctx.Value(usernameKey{}).(string)
	return username
}

// WithAccessToken stores the access token into the context.
func WithAccessToken(ctx context.Context, accessToken string) context.Context {
	return context.WithValue(ctx, accessTokenKey{}, accessToken)
}

// AccessToken retrieves the access token from the context.
func AccessToken(ctx context.Context) string {
	accessToken, _ := ctx.Value(accessTokenKey{}).(string)
	return accessToken
}

// WithRequestID stores the request ID into the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// RequestID retrieves the request ID from the context.
func RequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	return requestID
}

// WithTraceID stores the trace ID into the context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
    return context.WithValue(ctx, traceIDKey{}, traceID)
}

// TraceID retrieves the trace ID from the context.
func TraceID(ctx context.Context) string {
    traceID, _ := ctx.Value(traceIDKey{}).(string)
    return traceID
}

// WithLogger stores the logger into the context.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
    return context.WithValue(ctx, loggerKey{}, logger)
}

// Logger retrieves the logger from the context.
func Logger(ctx context.Context) *slog.Logger {
    logger, _ := ctx.Value(loggerKey{}).(*slog.Logger)
    return logger
}

// L is a short alias for Logger.
func L(ctx context.Context) *slog.Logger {
    return Logger(ctx)
}
