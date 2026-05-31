package cli

import (
	"testing"
)

func TestRequireFlag(t *testing.T) {
	err := RequireFlag("hello", "name")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	err = RequireFlag("", "name")
	if err == nil {
		t.Fatal("expected error for empty value")
	}
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if valErr.Field != "name" {
		t.Errorf("expected field 'name', got %q", valErr.Field)
	}

	// whitespace-only should also fail
	err = RequireFlag("   ", "name")
	if err == nil {
		t.Error("expected error for whitespace-only value")
	}
}

func TestRequireIntFlag(t *testing.T) {
	n, err := RequireIntFlag("42", "count")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if n != 42 {
		t.Errorf("expected 42, got %d", n)
	}

	_, err = RequireIntFlag("", "count")
	if err == nil {
		t.Fatal("expected error for empty value")
	}

	_, err = RequireIntFlag("abc", "count")
	if err == nil {
		t.Fatal("expected error for non-integer")
	}
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if valErr.Field != "count" {
		t.Errorf("expected field 'count', got %q", valErr.Field)
	}
}

func TestRequireFloatFlag(t *testing.T) {
	n, err := RequireFloatFlag("3.14", "price")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if n != 3.14 {
		t.Errorf("expected 3.14, got %f", n)
	}

	_, err = RequireFloatFlag("", "price")
	if err == nil {
		t.Fatal("expected error for empty value")
	}

	_, err = RequireFloatFlag("not-a-number", "price")
	if err == nil {
		t.Fatal("expected error for non-numeric")
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		url      string
		valid    bool
	}{
		{"http://localhost:8080", true},
		{"https://example.com", true},
		{"https://api.example.com/v1", true},
		{"", false},
		{"not-a-url", false},
		{"ftp://example.com", false},
		{"http://", false},
		{"://host", false},
	}
	for _, tt := range tests {
		err := ValidateURL(tt.url)
		if tt.valid && err != nil {
			t.Errorf("ValidateURL(%q) returned error: %v", tt.url, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("ValidateURL(%q) should have failed", tt.url)
		}
	}
}

func TestValidateEnum(t *testing.T) {
	allowed := []string{"table", "json", "yaml"}

	err := ValidateEnum("json", allowed, "output")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	err = ValidateEnum("xml", allowed, "output")
	if err == nil {
		t.Fatal("expected error for invalid enum value")
	}
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if valErr.Field != "output" {
		t.Errorf("expected field 'output', got %q", valErr.Field)
	}
}

func TestCollectErrors(t *testing.T) {
	err := CollectErrors()
	if err != nil {
		t.Errorf("expected nil for no errors, got %v", err)
	}

	err = CollectErrors(nil, nil)
	if err != nil {
		t.Errorf("expected nil for nil errors, got %v", err)
	}

	e1 := RequireFlag("", "name")
	e2 := ValidateURL("bad")
	err = CollectErrors(e1, e2)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if len(msg) == 0 {
		t.Error("expected non-empty error message")
	}
}

func TestValidationErrorMessage(t *testing.T) {
	err := &ValidationError{
		Field:   "name",
		Message: "--name is required",
	}
	expected := "invalid name: --name is required"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}
