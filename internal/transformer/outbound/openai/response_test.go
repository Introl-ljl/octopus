package openai

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	inboundanthropic "github.com/bestruirui/octopus/internal/transformer/inbound/anthropic"
)

func TestResponseOutboundTransformRequestPromotesAnthropicUserIDToUser(t *testing.T) {
	body := []byte(`{
		"model":"claude-test",
		"max_tokens":32,
		"metadata":{"user_id":"user-123"},
		"messages":[{"role":"user","content":"hello"}]
	}`)

	internalReq, err := (&inboundanthropic.MessagesInbound{}).TransformRequest(context.Background(), body)
	if err != nil {
		t.Fatalf("TransformRequest() error = %v", err)
	}

	if internalReq.User == nil || *internalReq.User != "user-123" {
		t.Fatalf("internalReq.User = %v, want user-123", internalReq.User)
	}
	if len(internalReq.Metadata) != 0 {
		t.Fatalf("internalReq.Metadata = %#v, want empty", internalReq.Metadata)
	}

	httpReq, err := (&ResponseOutbound{}).TransformRequest(context.Background(), internalReq, "https://example.com/v1", "test-key")
	if err != nil {
		t.Fatalf("TransformRequest() error = %v", err)
	}

	payloadBytes, err := io.ReadAll(httpReq.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got, ok := payload["user"].(string); !ok || got != "user-123" {
		t.Fatalf("payload user = %#v, want user-123", payload["user"])
	}
	if _, ok := payload["metadata"]; ok {
		t.Fatalf("payload metadata = %#v, want omitted", payload["metadata"])
	}
}
