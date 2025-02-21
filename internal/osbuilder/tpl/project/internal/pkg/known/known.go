package known

// Define HTTP/gRPC headers.
// gRPC uses HTTP/2 as its underlying transport protocol, and the HTTP/2 specification
// requires header keys to be in lowercase. Therefore, in gRPC, all header keys are
// forcibly converted to lowercase to conform to HTTP/2 requirements.
// In HTTP/1.x, many implementations preserve the case format set by the user,
// but some HTTP frameworks or tool libraries (such as certain web servers or proxies)
// may automatically convert headers to lowercase to simplify processing logic.
// For compatibility, all headers are uniformly set to lowercase here.
// Additionally, header keys prefixed with "x-" indicate they are custom headers.
const (
	// XRequestID defines the context key that represents the request ID.
	XRequestID = "x-request-id"

	// XUserID defines the context key that represents the ID of the requesting user.
	// UserID is unique throughout the user's entire lifecycle.
	XUserID = "x-user-id"

	// XUsername defines the context key that represents the requesting username.
	XUsername = "x-username"
)

// Define other constants.
const (
	// AdminUsername represents the username of the admin user.
	AdminUsername = "root"

	// MaxErrGroupConcurrency defines the maximum number of concurrent tasks for errgroup.
	// It is used to limit the number of simultaneous Goroutines executing within an errgroup,
	// preventing resource exhaustion and enhancing program stability.
	// This value can be adjusted based on the specific scenario and needs.
	MaxErrGroupConcurrency = 1000
)
