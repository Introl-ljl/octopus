package cli

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Message)
}

func RequireFlag(value string, flagName string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   flagName,
			Message: fmt.Sprintf("--%s is required", flagName),
		}
	}
	return nil
}

func RequireIntFlag(value string, flagName string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return 0, &ValidationError{
			Field:   flagName,
			Message: fmt.Sprintf("--%s is required", flagName),
		}
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, &ValidationError{
			Field:   flagName,
			Message: fmt.Sprintf("--%s must be a valid integer", flagName),
		}
	}
	return n, nil
}

func RequireFloatFlag(value string, flagName string) (float64, error) {
	if strings.TrimSpace(value) == "" {
		return 0, &ValidationError{
			Field:   flagName,
			Message: fmt.Sprintf("--%s is required", flagName),
		}
	}
	n, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, &ValidationError{
			Field:   flagName,
			Message: fmt.Sprintf("--%s must be a valid number", flagName),
		}
	}
	return n, nil
}

func ValidateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return &ValidationError{
			Field:   "url",
			Message: fmt.Sprintf("invalid URL: %v", err),
		}
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return &ValidationError{
			Field:   "url",
			Message: "URL must start with http:// or https://",
		}
	}
	if u.Host == "" {
		return &ValidationError{
			Field:   "url",
			Message: "URL must have a host",
		}
	}
	return nil
}

func ValidateEnum(value string, allowed []string, flagName string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return &ValidationError{
		Field:   flagName,
		Message: fmt.Sprintf("--%s must be one of: %s", flagName, strings.Join(allowed, ", ")),
	}
}

func CollectErrors(errs ...error) error {
	var msgs []string
	for _, e := range errs {
		if e != nil {
			msgs = append(msgs, e.Error())
		}
	}
	if len(msgs) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(msgs, "; "))
	}
	return nil
}
