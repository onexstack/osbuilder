/*
Package contextx provides extended functionality for contexts, allowing storage and extraction of user-related information such as user ID, username, and access tokens in a context.

The suffix "x" denotes extension or variation, making the package name concise and easy to remember. Functions within this package simplify the process of passing and managing user information in a context, suitable for scenarios where context-based data transfer is needed.

Typical usage:
In HTTP request middleware or service functions, these methods can be used to store user information in the context, ensuring safe sharing throughout the request lifecycle while avoiding the use of global variables and parameter passing.

Example:

	// Create a new context
	ctx := context.Background()

	// Store user ID and username in the context
	ctx = contextx.WithUserID(ctx, "user-xxxx")
	ctx = contextx.WithUsername(ctx, "sampleUser")

	// Extract user information from the context
	userID := contextx.UserID(ctx)
	username := contextx.Username(ctx)
*/
package contextx // import "{{.D.ModuleName}}/internal/pkg/contextx"
