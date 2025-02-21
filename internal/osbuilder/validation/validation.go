package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ValidateModulePath validates whether the given directory conforms to Go module naming conventions
func ValidateModulePath(path string) error {
	// If the path is empty, return an error immediately
	if path == "" {
		return errors.New("Module path cannot be empty")
	}

	// Check if the path begins with a domain name, e.g., github.com/
	if !strings.Contains(path, ".") || !strings.Contains(path, "/") {
		return errors.New("Module path must contain a valid domain name and be separated with '/'")
	}

	// Validate whether the path conforms to Go Modules naming conventions
	var modulePathRegex = `^([a-zA-Z0-9\-]+\.)+[a-zA-Z0-9\-]+(/[a-zA-Z0-9_\-]+)*$`
	matched, err := regexp.MatchString(modulePathRegex, path)
	if err != nil {
		return fmt.Errorf("Regex matching error: %v", err)
	}

	if !matched {
		return errors.New("Module path contains invalid characters or is in an incorrect format")
	}

	// Check if the path ends with `/` or `.`
	if strings.HasSuffix(path, "/") || strings.HasSuffix(path, ".") {
		return errors.New("Module path cannot end with '/' or '.'")
	}

	// The module path conforms to the convention
	return nil
}
