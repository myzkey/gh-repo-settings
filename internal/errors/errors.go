package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors
var (
	ErrConfigNotFound     = errors.New("configuration not found")
	ErrInvalidConfig      = errors.New("invalid configuration")
	ErrRepoNotFound       = errors.New("repository not found")
	ErrBranchNotFound     = errors.New("branch not found")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrRateLimited        = errors.New("rate limit exceeded")
	ErrNetworkError       = errors.New("network error")
	ErrSecretMissing      = errors.New("required secret is missing")
	ErrVariableMissing    = errors.New("required variable is missing")
	ErrBranchNotProtected = errors.New("branch protection not enabled")
	ErrPagesNotEnabled    = errors.New("GitHub Pages not enabled")
)

// ConfigError represents a configuration error
type ConfigError struct {
	File    string
	Message string
	Err     error
}

func (e *ConfigError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("config error in %s: %s", e.File, e.Message)
	}
	return fmt.Sprintf("config error: %s", e.Message)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new ConfigError
func NewConfigError(file, message string, err error) *ConfigError {
	return &ConfigError{
		File:    file,
		Message: message,
		Err:     err,
	}
}

// APIError represents a GitHub API error
type APIError struct {
	Endpoint   string
	Method     string
	StatusCode int
	Message    string
	Err        error
}

func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("API error: %s %s returned %d: %s", e.Method, e.Endpoint, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("API error: %s %s: %s", e.Method, e.Endpoint, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// NewAPIError creates a new APIError
func NewAPIError(method, endpoint string, statusCode int, message string, err error) *APIError {
	return &APIError{
		Endpoint:   endpoint,
		Method:     method,
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// Is checks if err matches target using errors.Is
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As checks if err can be assigned to target using errors.As
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
