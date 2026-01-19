package known

// Package known provides a collection of well-known and commonly used constants
// throughout the application, including HTTP/gRPC header keys and general
// configuration values.

// headerKey is an unexported type for HTTP/gRPC header keys to prevent accidental
// misuse of arbitrary strings as header names.
type headerKey string

const (
	// XRequestID represents the custom HTTP/gRPC header key for a request identifier.
	// Its value is "x-request-id".
	XRequestID headerKey = "x-request-id"

	// XUserID represents the custom HTTP/gRPC header key for the ID of the
	// requesting user. This ID is expected to be unique throughout the user's lifecycle.
	// Its value is "x-user-id".
	XUserID headerKey = "x-user-id"

	// XUsername represents the custom HTTP/gRPC header key for the requesting username.
	// Its value is "x-username".
	XUsername headerKey = "x-username"
)

// Application-specific constants.
const (
	// AdminUsername represents the default username for an administrative user.
	// Its value is "root".
	AdminUsername = "root"

	// MaxErrGroupConcurrency defines the maximum number of concurrent tasks
	// allowed within an errgroup. This limits simultaneous Goroutine execution
	// to prevent resource exhaustion and enhance program stability.
	// The default value is 1000, which can be adjusted based on specific
	// scenario and performance requirements.
	MaxErrGroupConcurrency = 1000
)
