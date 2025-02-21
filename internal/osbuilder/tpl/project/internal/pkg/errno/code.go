package errno

import (
	"net/http"

	"github.com/onexstack/onexstack/pkg/errorsx"
)

var (
	// OK represents a successful request.
	OK = &errorsx.ErrorX{Code: http.StatusOK, Message: ""}

	// ErrInternal represents all unknown server-side errors.
	ErrInternal = errorsx.ErrInternal

	// ErrNotFound indicates that the resource was not found.
	ErrNotFound = errorsx.ErrNotFound

	// ErrBind indicates an error occurred while binding the request body.
	ErrBind = errorsx.ErrBind

	// ErrInvalidArgument indicates that argument validation failed.
	ErrInvalidArgument = errorsx.ErrInvalidArgument

	// ErrUnauthenticated indicates authentication failure.
	ErrUnauthenticated = errorsx.ErrUnauthenticated

	// ErrPermissionDenied indicates the request was forbidden due to insufficient permissions.
	ErrPermissionDenied = errorsx.ErrPermissionDenied

	// ErrOperationFailed indicates that the operation failed.
	ErrOperationFailed = errorsx.ErrOperationFailed

	// ErrPageNotFound indicates that the page was not found.
	ErrPageNotFound = &errorsx.ErrorX{Code: http.StatusNotFound, Reason: "NotFound.PageNotFound", Message: "Page not found."}

	// ErrSignToken indicates an error occurred while signing a JWT token.
	ErrSignToken = &errorsx.ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated.SignToken", Message: "Error occurred while signing the JSON web token."}

	// ErrTokenInvalid indicates that the JWT token format was invalid.
	ErrTokenInvalid = &errorsx.ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated.TokenInvalid", Message: "Token was invalid."}

	// ErrDBRead indicates a database read failure.
	ErrDBRead = &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.DBRead", Message: "Database read failure."}

	// ErrDBWrite indicates a database write failure.
	ErrDBWrite = &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.DBWrite", Message: "Database write failure."}

	// ErrAddRole indicates an error occurred while adding a role.
	ErrAddRole = &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.AddRole", Message: "Error occurred while adding the role."}

	// ErrRemoveRole indicates an error occurred while removing a role.
	ErrRemoveRole = &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.RemoveRole", Message: "Error occurred while removing the role."}
)
