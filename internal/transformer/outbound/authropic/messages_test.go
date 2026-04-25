package authropic

import (
	"testing"

	"github.com/bestruirui/octopus/internal/transformer/model"
)

func TestConvertToAnthropicRequestUsesUserFieldForMetadata(t *testing.T) {
	userID := "user-123"
	req := &model.InternalLLMRequest{
		Model: "claude-test",
		User:  &userID,
	}

	result := convertToAnthropicRequest(req)
	if result.Metadata == nil {
		t.Fatal("result.Metadata = nil, want user metadata")
	}
	if result.Metadata.UserID != userID {
		t.Fatalf("result.Metadata.UserID = %q, want %q", result.Metadata.UserID, userID)
	}
}
