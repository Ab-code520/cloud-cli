package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidationError represents a validation failure.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// Validator provides a fluent API for input validation.
type Validator struct {
	errors []ValidationError
}

// NewValidator creates a new validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Required checks that a string is not empty.
func (v *Validator) Required(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: "is required",
		})
	}
	return v
}

// NonEmptyPath checks that a path is valid and not empty.
func (v *Validator) NonEmptyPath(field, path string) *Validator {
	if strings.TrimSpace(path) == "" {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: "path is required",
		})
		return v
	}

	// Check for null bytes
	if strings.ContainsRune(path, 0x00) {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: "path contains null bytes",
		})
	}

	// Clean and validate path separators
	cleaned := filepath.Clean(path)
	if cleaned == "." {
		// Root path is valid
	}

	return v
}

// FileExists checks that a local file exists and is readable.
func (v *Validator) FileExists(field, path string) *Validator {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			v.errors = append(v.errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("file not found: %s", path),
			})
		} else {
			v.errors = append(v.errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("cannot access file: %s", err),
			})
		}
		return v
	}

	if info.IsDir() {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: "expected a file, got a directory",
		})
	}

	return v
}

// RangeInt checks that an integer is within a range.
func (v *Validator) RangeInt(field string, value, min, max int) *Validator {
	if value < min || value > max {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be between %d and %d, got %d", min, max, value),
		})
	}
	return v
}

// OneOf checks that a value is one of the allowed values.
func (v *Validator) OneOf(field, value string, allowed ...string) *Validator {
	for _, a := range allowed {
		if value == a {
			return v
		}
	}
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")),
	})
	return v
}

// Validate returns the first validation error, or nil if all passed.
func (v *Validator) Validate() error {
	if len(v.errors) > 0 {
		return &v.errors[0]
	}
	return nil
}

// ValidateAll returns all validation errors combined.
func (v *Validator) ValidateAll() error {
	if len(v.errors) == 0 {
		return nil
	}
	if len(v.errors) == 1 {
		return &v.errors[0]
	}
	msgs := make([]string, len(v.errors))
	for i, e := range v.errors {
		msgs[i] = e.Error()
	}
	return fmt.Errorf("multiple validation errors:\n  - %s", strings.Join(msgs, "\n  - "))
}
