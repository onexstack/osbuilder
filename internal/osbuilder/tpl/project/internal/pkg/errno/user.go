package errno

import (
	"net/http"

	"github.com/onexstack/onexstack/pkg/errorsx"
)

var (
	// ErrUsernameInvalid indicates that the username is invalid.
	ErrUsernameInvalid = errorsx.New(
		http.StatusBadRequest,
		"InvalidArgument.UsernameInvalid",
		"Username must consist of letters, digits, and underscores only, and its length must be between 3 and 20 characters.",
	)

	// ErrPasswordInvalid indicates that the password is invalid.
	ErrPasswordInvalid = errorsx.New(http.StatusBadRequest, "InvalidArgument.PasswordInvalid", "Password is incorrect.")

	// ErrUserAlreadyExists indicates that the user already exists.
	ErrUserAlreadyExists = errorsx.New(http.StatusBadRequest, "AlreadyExist.UserAlreadyExists", "User already exists.")

	// ErrUserNotFound indicates that the specified user was not found.
	ErrUserNotFound = errorsx.New(http.StatusNotFound, "NotFound.UserNotFound", "User not found.")
)
