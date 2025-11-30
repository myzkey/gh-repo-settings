package errors

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrConfigNotFound", ErrConfigNotFound, "configuration not found"},
		{"ErrInvalidConfig", ErrInvalidConfig, "invalid configuration"},
		{"ErrRepoNotFound", ErrRepoNotFound, "repository not found"},
		{"ErrBranchNotFound", ErrBranchNotFound, "branch not found"},
		{"ErrPermissionDenied", ErrPermissionDenied, "permission denied"},
		{"ErrRateLimited", ErrRateLimited, "rate limit exceeded"},
		{"ErrNetworkError", ErrNetworkError, "network error"},
		{"ErrSecretMissing", ErrSecretMissing, "required secret is missing"},
		{"ErrVariableMissing", ErrVariableMissing, "required variable is missing"},
		{"ErrBranchNotProtected", ErrBranchNotProtected, "branch protection not enabled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("got %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestConfigError(t *testing.T) {
	t.Run("with file", func(t *testing.T) {
		err := &ConfigError{
			File:    "config.yaml",
			Message: "invalid syntax",
			Err:     ErrInvalidConfig,
		}

		want := "config error in config.yaml: invalid syntax"
		if err.Error() != want {
			t.Errorf("got %q, want %q", err.Error(), want)
		}

		if err.Unwrap() != ErrInvalidConfig {
			t.Errorf("Unwrap() should return wrapped error")
		}
	})

	t.Run("without file", func(t *testing.T) {
		err := &ConfigError{
			Message: "missing required field",
		}

		want := "config error: missing required field"
		if err.Error() != want {
			t.Errorf("got %q, want %q", err.Error(), want)
		}
	})
}

func TestNewConfigError(t *testing.T) {
	wrapped := errors.New("underlying error")
	err := NewConfigError("test.yaml", "parse failed", wrapped)

	if err.File != "test.yaml" {
		t.Errorf("File = %q, want %q", err.File, "test.yaml")
	}
	if err.Message != "parse failed" {
		t.Errorf("Message = %q, want %q", err.Message, "parse failed")
	}
	if err.Err != wrapped {
		t.Errorf("Err should be the wrapped error")
	}
}

func TestAPIError(t *testing.T) {
	t.Run("with status code", func(t *testing.T) {
		err := &APIError{
			Endpoint:   "/repos/owner/repo",
			Method:     "GET",
			StatusCode: 404,
			Message:    "Not Found",
			Err:        ErrRepoNotFound,
		}

		want := "API error: GET /repos/owner/repo returned 404: Not Found"
		if err.Error() != want {
			t.Errorf("got %q, want %q", err.Error(), want)
		}

		if err.Unwrap() != ErrRepoNotFound {
			t.Errorf("Unwrap() should return wrapped error")
		}
	})

	t.Run("without status code", func(t *testing.T) {
		err := &APIError{
			Endpoint: "/repos/owner/repo",
			Method:   "POST",
			Message:  "connection refused",
		}

		want := "API error: POST /repos/owner/repo: connection refused"
		if err.Error() != want {
			t.Errorf("got %q, want %q", err.Error(), want)
		}
	})
}

func TestNewAPIError(t *testing.T) {
	wrapped := errors.New("timeout")
	err := NewAPIError("GET", "/api/test", 500, "Internal Server Error", wrapped)

	if err.Method != "GET" {
		t.Errorf("Method = %q, want %q", err.Method, "GET")
	}
	if err.Endpoint != "/api/test" {
		t.Errorf("Endpoint = %q, want %q", err.Endpoint, "/api/test")
	}
	if err.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want %d", err.StatusCode, 500)
	}
	if err.Message != "Internal Server Error" {
		t.Errorf("Message = %q, want %q", err.Message, "Internal Server Error")
	}
	if err.Err != wrapped {
		t.Errorf("Err should be the wrapped error")
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "repo.visibility",
		Message: "must be public, private, or internal",
	}

	want := "validation error: repo.visibility: must be public, private, or internal"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("labels.color", "invalid hex color")

	if err.Field != "labels.color" {
		t.Errorf("Field = %q, want %q", err.Field, "labels.color")
	}
	if err.Message != "invalid hex color" {
		t.Errorf("Message = %q, want %q", err.Message, "invalid hex color")
	}
}

func TestIs(t *testing.T) {
	err := &ConfigError{
		File:    "test.yaml",
		Message: "invalid",
		Err:     ErrInvalidConfig,
	}

	if !Is(err, ErrInvalidConfig) {
		t.Error("Is() should return true for wrapped error")
	}

	if Is(err, ErrConfigNotFound) {
		t.Error("Is() should return false for different error")
	}
}

func TestAs(t *testing.T) {
	t.Run("ConfigError", func(t *testing.T) {
		err := error(NewConfigError("test.yaml", "failed", nil))

		var configErr *ConfigError
		if !As(err, &configErr) {
			t.Error("As() should return true for ConfigError")
		}
		if configErr.File != "test.yaml" {
			t.Errorf("File = %q, want %q", configErr.File, "test.yaml")
		}
	})

	t.Run("APIError", func(t *testing.T) {
		err := error(NewAPIError("GET", "/test", 404, "Not Found", nil))

		var apiErr *APIError
		if !As(err, &apiErr) {
			t.Error("As() should return true for APIError")
		}
		if apiErr.StatusCode != 404 {
			t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, 404)
		}
	})

	t.Run("ValidationError", func(t *testing.T) {
		err := error(NewValidationError("field", "invalid"))

		var valErr *ValidationError
		if !As(err, &valErr) {
			t.Error("As() should return true for ValidationError")
		}
		if valErr.Field != "field" {
			t.Errorf("Field = %q, want %q", valErr.Field, "field")
		}
	})

	t.Run("non-matching type", func(t *testing.T) {
		err := errors.New("simple error")

		var configErr *ConfigError
		if As(err, &configErr) {
			t.Error("As() should return false for non-matching type")
		}
	})
}

func TestErrorChaining(t *testing.T) {
	// Test error chain: ValidationError -> ConfigError -> ErrInvalidConfig
	valErr := NewValidationError("repo.name", "required")
	configErr := NewConfigError("config.yaml", valErr.Error(), ErrInvalidConfig)

	// Check the chain works with Is
	if !Is(configErr, ErrInvalidConfig) {
		t.Error("Is() should find ErrInvalidConfig in chain")
	}

	// Check the chain works with As
	var foundConfig *ConfigError
	if !As(configErr, &foundConfig) {
		t.Error("As() should find ConfigError")
	}
}
