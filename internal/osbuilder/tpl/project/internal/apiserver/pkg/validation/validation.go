package validation

import (
	{{- if .Web.WithUser }}
	"regexp"

	{{- end}}
	"github.com/google/wire"

	"{{.D.ModuleName}}/internal/{{.Web.Name}}/store"
	{{- if .Web.WithUser }}
	"{{.D.ModuleName}}/internal/pkg/errno"
	{{- end}}
)

// Validator handles custom business validation logic.
// It holds dependencies required for deep validation, such as database access.
type Validator struct {
	// Some complex validation logic may require direct database queries.
	// This is just an example. If validation requires other dependencies 
	// like clients, services, resources, etc., they can all be injected here.
	store store.IStore
}

{{- if .Web.WithUser }}
// Use globally precompiled regular expressions to avoid creating and compiling them repeatedly.
var (
	lengthRegex = regexp.MustCompile(`^.{3,20}$`)                                        // Length between 3 and 20 characters
	validRegex  = regexp.MustCompile(`^[A-Za-z0-9_]+$`)                                  // Only letters, numbers, and underscores
	letterRegex = regexp.MustCompile(`[A-Za-z]`)                                         // At least one letter
	numberRegex = regexp.MustCompile(`\d`)                                               // At least one number
	emailRegex  = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`) // Email format
	phoneRegex  = regexp.MustCompile(`^1[3-9]\d{9}$`)                                    // Chinese phone number
)
{{- end}}

// ProviderSet is the Wire provider set for the validation package.
var ProviderSet = wire.NewSet(New, wire.Bind(new(any), new(*Validator)))

// New creates and initializes a new Validator instance with the required dependencies.
func New(ds store.IStore) *Validator {
	return &Validator{store: ds}
}

{{- if .Web.WithUser }}
// isValidUsername validates if a username is valid.
func isValidUsername(username string) bool {
	// Validate length
	if !lengthRegex.MatchString(username) {
		return false
	}
	// Validate character legality
	if !validRegex.MatchString(username) {
		return false
	}
	return true
}

// isValidPassword checks whether a password meets complexity requirements.
func isValidPassword(password string) error {
	switch {
	// Check if the new password is empty
	case password == "":
		return errno.ErrInvalidArgument.WithMessage("password cannot be empty")
	// Check the length requirement of the new password
	case len(password) < 6:
		return errno.ErrInvalidArgument.WithMessage("password must be at least 6 characters long")
	// Use a regular expression to check if it contains at least one letter
	case !letterRegex.MatchString(password):
		return errno.ErrInvalidArgument.WithMessage("password must contain at least one letter")
	// Use a regular expression to check if it contains at least one number
	case !numberRegex.MatchString(password):
		return errno.ErrInvalidArgument.WithMessage("password must contain at least one number")
	}
	return nil
}

// isValidEmail checks whether an email is valid.
func isValidEmail(email string) error {
	// Check if the email is empty
	if email == "" {
		return errno.ErrInvalidArgument.WithMessage("email cannot be empty")
	}

	// Validate email format using a regular expression
	if !emailRegex.MatchString(email) {
		return errno.ErrInvalidArgument.WithMessage("invalid email format")
	}

	return nil
}

// isValidPhone checks whether a phone number is valid.
func isValidPhone(phone string) error {
	// Check if the phone number is empty
	if phone == "" {
		return errno.ErrInvalidArgument.WithMessage("phone cannot be empty")
	}

	// Validate the phone number format (assumed to be a Chinese phone number, 11 digits)
	if !phoneRegex.MatchString(phone) {
		return errno.ErrInvalidArgument.WithMessage("invalid phone format")
	}

	return nil
}
{{- end }}
