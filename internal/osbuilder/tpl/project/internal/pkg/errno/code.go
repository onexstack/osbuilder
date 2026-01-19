package errno

import (
	"net/http"

	"github.com/onexstack/onexstack/pkg/errorsx"
)

// Global application-specific error definitions.
// These errors are intended to be returned when specific conditions occur.

// OK represents a successful request.
var OK = errorsx.New(http.StatusOK, "", "")

// ErrInternal represents all unknown server-side errors.
// This error is typically used for unexpected issues within the service.
var ErrInternal = errorsx.ErrInternal // Assuming errorsx provides this common error

// ErrNotFound indicates that the requested resource was not found.
var ErrNotFound = errorsx.ErrNotFound // Assuming errorsx provides this common error

// ErrBind indicates an error occurred while binding the request body.
// This typically signifies malformed input from the client.
var ErrBind = errorsx.ErrBind // Assuming errorsx provides this common error

// ErrInvalidArgument indicates that an argument provided in the request
// failed validation.
var ErrInvalidArgument = errorsx.ErrInvalidArgument // Assuming errorsx provides this common error

// ErrUnauthenticated indicates authentication failure.
// This error is returned when the client lacks valid authentication credentials.
var ErrUnauthenticated = errorsx.ErrUnauthenticated // Assuming errorsx provides this common error

// ErrPermissionDenied indicates that the request was forbidden due to
// insufficient permissions for the authenticated user.
var ErrPermissionDenied = errorsx.ErrPermissionDenied // Assuming errorsx provides this common error

// ErrOperationFailed indicates that a generic operation failed for an
// unspecified reason. More specific errors should be used when possible.
var ErrOperationFailed = errorsx.ErrOperationFailed // Assuming errorsx provides this common error

// ErrPageNotFound indicates that the specific page or route was not found.
var ErrPageNotFound = errorsx.New(http.StatusNotFound, "NotFound.PageNotFound", "The requested page was not found.")

// ErrSignToken indicates an error occurred during the process of signing
// a JSON Web Token (JWT).
var ErrSignToken = errorsx.New(http.StatusUnauthorized, "Unauthenticated.SignToken", "Error occurred while signing the JSON web token.")

// ErrTokenInvalid indicates that the provided JWT token was invalid,
// either due to malformation, expiration, or incorrect signature.
var ErrTokenInvalid = errorsx.New(http.StatusUnauthorized, "Unauthenticated.TokenInvalid", "The provided token was invalid.")

// ErrDBRead indicates a failure during a database read operation.
var ErrDBRead = errorsx.New(http.StatusInternalServerError, "InternalError.DBRead", "A database read operation failed.")

// ErrDBWrite indicates a failure during a database write operation.
var ErrDBWrite = errorsx.New(http.StatusInternalServerError, "InternalError.DBWrite", "A database write operation failed.")

// ErrAddRole indicates an error occurred while adding a user role.
var ErrAddRole = errorsx.New(http.StatusInternalServerError, "InternalError.AddRole", "Error occurred while adding the role.")

// ErrRemoveRole indicates an error occurred while removing a user role.
var ErrRemoveRole = errorsx.New(http.StatusInternalServerError, "InternalError.RemoveRole", "Error occurred while removing the role.")
