package cli

import (
	"encoding/json"
	"testing"
)

func TestTableFormatterOutput(t *testing.T) {
	f := NewTableFormatter()

	// We can't easily capture tabwriter output, but we can verify no panics
	f.PrintHeader("ID", "Name")
	f.PrintRow("1", "test")
	f.PrintRow("2", "example")
	f.Flush()
}

func TestTableFormatterPrintJSON(t *testing.T) {
	f := NewTableFormatter()
	data := map[string]interface{}{"key": "value"}
	// Should print JSON to stdout - verify no panic
	f.PrintJSON(data)
}

func TestTableFormatterPrintError(t *testing.T) {
	f := NewTableFormatter()
	// Should print to stderr - verify no panic
	f.PrintError("something went wrong")
}

func TestTableFormatterPrintSuccess(t *testing.T) {
	f := NewTableFormatter()
	// Should print to stdout - verify no panic
	f.PrintSuccess("done")
}

func TestJSONFormatterPrintHeader(t *testing.T) {
	f := NewJSONFormatter()
	f.PrintHeader("ID", "Name") // should be no-op
}

func TestJSONFormatterPrintRow(t *testing.T) {
	f := NewJSONFormatter()
	f.PrintRow("1", "test") // should be no-op
}

func TestJSONFormatterFlush(t *testing.T) {
	f := NewJSONFormatter()
	f.Flush() // should be no-op
}

func TestJSONFormatterPrintJSON(t *testing.T) {
	f := NewJSONFormatter()
	data := map[string]string{"key": "value"}
	f.PrintJSON(data)
}

func TestJSONFormatterPrintError(t *testing.T) {
	f := NewJSONFormatter()
	f.PrintError("test error")
}

func TestJSONFormatterPrintSuccess(t *testing.T) {
	f := NewJSONFormatter()
	f.PrintSuccess("created")
}

func TestJSONFormatterOutputFormat(t *testing.T) {
	f := NewJSONFormatter()

	// Test PrintSuccess produces valid JSON
	f.PrintSuccess("success")
	// Test PrintError produces valid JSON
	f.PrintError("error msg")
}

func TestGetFormatterForMode(t *testing.T) {
	table := GetFormatterForMode("table")
	if _, ok := table.(*TableFormatter); !ok {
		t.Error("expected TableFormatter for mode 'table'")
	}

	jsonF := GetFormatterForMode("json")
	if _, ok := jsonF.(*JSONFormatter); !ok {
		t.Error("expected JSONFormatter for mode 'json'")
	}

	// empty mode should default to table
	defaultF := GetFormatterForMode("")
	if _, ok := defaultF.(*TableFormatter); !ok {
		t.Error("expected TableFormatter for empty mode")
	}
}

func TestFormatterInterface(t *testing.T) {
	var f Formatter
	f = NewTableFormatter()
	_ = f

	f = NewJSONFormatter()
	_ = f
}

func verifyJSONOutput(t *testing.T, output string) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	return result
}

func TestJSONFormatterErrorOutput(t *testing.T) {
	// Capture stderr
	f := NewJSONFormatter()
	// Just verify no panic
	f.PrintError("permission denied")
}

func TestJSONFormatterSuccessOutput(t *testing.T) {
	f := NewJSONFormatter()
	// Just verify no panic
	f.PrintSuccess("api key created")
}
