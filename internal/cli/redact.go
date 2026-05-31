package cli

import "strings"

type RedactRule struct {
	FieldNames []string
}

var defaultRedactFields = []string{
	"token", "api_key", "channel_key", "api-key",
	"authorization", "secret", "password", "key",
}

func RedactFields(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(data))
	for k, v := range data {
		if isRedactField(k) {
			result[k] = redactValue(v)
		} else if nested, ok := v.(map[string]interface{}); ok {
			result[k] = RedactFields(nested)
		} else if arr, ok := v.([]interface{}); ok {
			result[k] = redactArray(arr)
		} else {
			result[k] = v
		}
	}
	return result
}

func redactArray(arr []interface{}) []interface{} {
	result := make([]interface{}, len(arr))
	for i, v := range arr {
		if m, ok := v.(map[string]interface{}); ok {
			result[i] = RedactFields(m)
		} else {
			result[i] = v
		}
	}
	return result
}

func isRedactField(name string) bool {
	lower := strings.ToLower(name)
	for _, f := range defaultRedactFields {
		if lower == f || strings.HasSuffix(lower, "_"+f) {
			return true
		}
	}
	return false
}

func redactValue(v interface{}) string {
	s, ok := v.(string)
	if !ok || s == "" {
		return "***"
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "***" + s[len(s)-4:]
}

func IsRedactedValue(s string) bool {
	return s == "***" || strings.Contains(s, "***")
}
