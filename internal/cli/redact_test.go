package cli

import (
	"fmt"
	"testing"
)

func TestRedactToken(t *testing.T) {
	data := map[string]interface{}{
		"token": "sk-abcdefghijklmnopqrstuvwxyz",
	}
	redacted := RedactFields(data)
	val, ok := redacted["token"].(string)
	if !ok {
		t.Fatal("expected string value")
	}
	if val == "sk-abcdefghijklmnopqrstuvwxyz" {
		t.Error("token was not redacted")
	}
	if len(val) < 10 {
		t.Errorf("redacted value too short: %q", val)
	}
}

func TestRedactAPIKey(t *testing.T) {
	data := map[string]interface{}{
		"api_key": "secret-api-key-12345",
	}
	redacted := RedactFields(data)
	val, ok := redacted["api_key"].(string)
	if !ok {
		t.Fatal("expected string value")
	}
	if !IsRedactedValue(val) {
		t.Errorf("expected redacted value, got %q", val)
	}
}

func TestRedactPassword(t *testing.T) {
	data := map[string]interface{}{
		"password": "hunter2",
	}
	redacted := RedactFields(data)
	if redacted["password"] != "***" {
		t.Errorf("short password should be fully redacted, got %q", redacted["password"])
	}
}

func TestRedactShortValue(t *testing.T) {
	data := map[string]interface{}{
		"secret": "123",
	}
	redacted := RedactFields(data)
	if redacted["secret"] != "***" {
		t.Errorf("short value should be '***', got %q", redacted["secret"])
	}
}

func TestRedactNonStringValue(t *testing.T) {
	data := map[string]interface{}{
		"api_key": 12345,
	}
	redacted := RedactFields(data)
	if redacted["api_key"] != "***" {
		t.Errorf("non-string should be '***', got %q", redacted["api_key"])
	}
}

func TestRedactEmptyString(t *testing.T) {
	data := map[string]interface{}{
		"api_key": "",
	}
	redacted := RedactFields(data)
	if redacted["api_key"] != "***" {
		t.Errorf("empty string should be '***', got %q", redacted["api_key"])
	}
}

func TestRedactKeepNonSecret(t *testing.T) {
	data := map[string]interface{}{
		"name":    "test-channel",
		"enabled": true,
		"count":   42,
	}
	redacted := RedactFields(data)
	if redacted["name"] != "test-channel" {
		t.Errorf("name should not be redacted, got %q", redacted["name"])
	}
	if redacted["enabled"] != true {
		t.Error("enabled should not be redacted")
	}
	if redacted["count"] != 42 {
		t.Error("count should not be redacted")
	}
}

func TestRedactNestedMap(t *testing.T) {
	data := map[string]interface{}{
		"config": map[string]interface{}{
			"api_key": "nested-secret",
			"url":     "http://example.com",
		},
	}
	redacted := RedactFields(data)
	nested, ok := redacted["config"].(map[string]interface{})
	if !ok {
		t.Fatal("expected nested map")
	}
	if !IsRedactedValue(nested["api_key"].(string)) {
		t.Errorf("nested api_key should be redacted, got %q", nested["api_key"])
	}
	if nested["url"] != "http://example.com" {
		t.Errorf("nested url should be preserved, got %q", nested["url"])
	}
}

func TestRedactNestedArray(t *testing.T) {
	data := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"name":   "item1",
				"api_key": "key1",
			},
			map[string]interface{}{
				"name":   "item2",
				"api_key": "key2",
			},
		},
	}
	redacted := RedactFields(data)
	items, ok := redacted["items"].([]interface{})
	if !ok {
		t.Fatal("expected array")
	}
	for i, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			t.Fatalf("item %d: expected map", i)
		}
		if !IsRedactedValue(m["api_key"].(string)) {
			t.Errorf("item %d api_key should be redacted", i)
		}
		if m["name"] != fmt.Sprintf("item%d", i+1) {
			t.Errorf("item %d name should be preserved", i)
		}
	}
}

func TestIsRedactedValue(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"***", true},
		{"sk-***xyz", true},
		{"sk-abcdefgh", false},
		{"", false},
		{"hello", false},
	}
	for _, tt := range tests {
		result := IsRedactedValue(tt.input)
		if result != tt.expected {
			t.Errorf("IsRedactedValue(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestRedactSuffixMatch(t *testing.T) {
	data := map[string]interface{}{
		"channel_key": "should-redact",
		"authorization": "Bearer tok",
	}
	redacted := RedactFields(data)
	if !IsRedactedValue(redacted["channel_key"].(string)) {
		t.Error("channel_key should be redacted")
	}
	if !IsRedactedValue(redacted["authorization"].(string)) {
		t.Error("authorization should be redacted")
	}
}

func TestRedactPartialValue(t *testing.T) {
	data := map[string]interface{}{
		"token": "abcdefghijklmnop", // 16 chars
	}
	redacted := RedactFields(data)
	val := redacted["token"].(string)
	// Should be: abcd***mnop
	if len(val) != 11 {
		t.Errorf("expected 11 chars (4+3+4), got %d: %q", len(val), val)
	}
	if val[:4] != "abcd" {
		t.Errorf("expected prefix 'abcd', got %q", val[:4])
	}
	if val[len(val)-4:] != "mnop" {
		t.Errorf("expected suffix 'mnop', got %q", val[len(val)-4:])
	}
}
