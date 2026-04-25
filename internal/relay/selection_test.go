package relay

import (
	"testing"

	dbmodel "github.com/bestruirui/octopus/internal/model"
	"github.com/bestruirui/octopus/internal/transformer"
	"github.com/bestruirui/octopus/internal/transformer/outbound"
)

func outboundTypePtr(v outbound.OutboundType) *outbound.OutboundType {
	return &v
}

func TestPreferredOutboundTypeForFormat(t *testing.T) {
	tests := []struct {
		name string
		in   transformer.Format
		want *outbound.OutboundType
	}{
		{name: "chat", in: transformer.FormatOpenAIChat, want: outboundTypePtr(outbound.OutboundTypeOpenAIChat)},
		{name: "responses", in: transformer.FormatOpenAIResponse, want: outboundTypePtr(outbound.OutboundTypeOpenAIResponse)},
		{name: "embedding", in: transformer.FormatOpenAIEmbedding, want: outboundTypePtr(outbound.OutboundTypeOpenAIEmbedding)},
		{name: "anthropic", in: transformer.FormatAnthropic, want: outboundTypePtr(outbound.OutboundTypeAnthropic)},
		{name: "unknown", in: transformer.FormatGemini, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := preferredOutboundTypeForFormat(tt.in)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("preferredOutboundTypeForFormat(%s) = %v, want nil", tt.in, *got)
				}
				return
			}
			if got == nil || *got != *tt.want {
				t.Fatalf("preferredOutboundTypeForFormat(%s) = %v, want %v", tt.in, got, *tt.want)
			}
		})
	}
}

func TestChannelGetBaseUrlByType(t *testing.T) {
	channel := &dbmodel.Channel{
		Type: outbound.OutboundTypeOpenAIChat,
		BaseUrls: []dbmodel.BaseUrl{
			{URL: "https://chat.example.com", Delay: 90, Type: outboundTypePtr(outbound.OutboundTypeOpenAIChat)},
			{URL: "https://responses.example.com", Delay: 80, Type: outboundTypePtr(outbound.OutboundTypeOpenAIResponse)},
			{URL: "https://legacy.example.com", Delay: 10},
		},
	}

	baseURL, selectedType := channel.GetBaseUrlByType(outboundTypePtr(outbound.OutboundTypeOpenAIResponse))
	if baseURL != "https://responses.example.com" {
		t.Fatalf("responses baseURL = %q, want %q", baseURL, "https://responses.example.com")
	}
	if selectedType != outbound.OutboundTypeOpenAIResponse {
		t.Fatalf("responses selectedType = %v, want %v", selectedType, outbound.OutboundTypeOpenAIResponse)
	}

	baseURL, selectedType = channel.GetBaseUrlByType(outboundTypePtr(outbound.OutboundTypeAnthropic))
	if baseURL != "https://legacy.example.com" {
		t.Fatalf("fallback baseURL = %q, want %q", baseURL, "https://legacy.example.com")
	}
	if selectedType != outbound.OutboundTypeOpenAIChat {
		t.Fatalf("fallback selectedType = %v, want %v", selectedType, outbound.OutboundTypeOpenAIChat)
	}
}

