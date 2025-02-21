package contextx

import (
	"context"
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
