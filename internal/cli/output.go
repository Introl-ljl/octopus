package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

type OutputFormat int

const (
	OutputTable OutputFormat = iota
	OutputJSON
)

type Formatter interface {
	PrintHeader(headers ...string)
	PrintRow(fields ...string)
	Flush()
	PrintJSON(v interface{})
	PrintError(msg string)
	PrintSuccess(msg string)
}

type TableFormatter struct {
	writer *tabwriter.Writer
}

func NewTableFormatter() *TableFormatter {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	return &TableFormatter{writer: w}
}

func (t *TableFormatter) PrintHeader(headers ...string) {
	line := strings.Join(headers, "\t")
	fmt.Fprintln(t.writer, line)
	seps := make([]string, len(headers))
	for i, h := range headers {
		seps[i] = strings.Repeat("-", len(h))
	}
	fmt.Fprintln(t.writer, strings.Join(seps, "\t"))
}

func (t *TableFormatter) PrintRow(fields ...string) {
	fmt.Fprintln(t.writer, strings.Join(fields, "\t"))
}

func (t *TableFormatter) Flush() {
	t.writer.Flush()
}

func (t *TableFormatter) PrintJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func (t *TableFormatter) PrintError(msg string) {
	fmt.Fprintln(os.Stderr, "Error:", msg)
}

func (t *TableFormatter) PrintSuccess(msg string) {
	fmt.Println(msg)
}

type JSONFormatter struct{}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (j *JSONFormatter) PrintHeader(headers ...string) {}

func (j *JSONFormatter) PrintRow(fields ...string) {}

func (j *JSONFormatter) Flush() {}

func (j *JSONFormatter) PrintJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		j.PrintError(fmt.Sprintf("failed to marshal JSON: %v", err))
		return
	}
	fmt.Println(string(data))
}

func (j *JSONFormatter) PrintError(msg string) {
	errOutput := map[string]string{"error": msg}
	data, _ := json.Marshal(errOutput)
	fmt.Fprintln(os.Stderr, string(data))
}

func (j *JSONFormatter) PrintSuccess(msg string) {
	output := map[string]string{"message": msg}
	data, _ := json.Marshal(output)
	fmt.Println(string(data))
}

func GetFormatter() Formatter {
	profile := GetActiveProfile()
	if profile != nil && profile.DefaultOutput == "json" {
		return NewJSONFormatter()
	}
	return NewTableFormatter()
}

func GetFormatterForMode(mode string) Formatter {
	if mode == "json" {
		return NewJSONFormatter()
	}
	return NewTableFormatter()
}
