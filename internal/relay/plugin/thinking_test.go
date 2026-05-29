package plugin

import (
	"context"
	"strings"
	"testing"

	tmodel "github.com/bestruirui/octopus/internal/transformer/model"
)

func TestThinkingCacheIsScopedByCallerAndConversation(t *testing.T) {
	p := &DeepSeekThinkingPlugin{
		store: &reasoningStore{
			data: make(map[string]*reasoningEntry),
		},
	}

	userContent := "what is next?"
	assistantContent := "the visible answer"
	reasoningContent := "private reasoning"

	storeReq := &tmodel.InternalLLMRequest{
		Model:               "deepseek-reasoner",
		ReasoningCacheScope: "api_key:1",
		Messages: []tmodel.Message{{
			Role:    "user",
			Content: tmodel.MessageContent{Content: &userContent},
		}},
	}
	resp := &tmodel.InternalLLMResponse{
		Model: "deepseek-reasoner",
		Choices: []tmodel.Choice{{
			Message: &tmodel.Message{
				Role:             "assistant",
				Content:          tmodel.MessageContent{Content: &assistantContent},
				ReasoningContent: &reasoningContent,
			},
		}},
	}
	if err := p.OnResponse(context.Background(), storeReq, resp); err != nil {
		t.Fatalf("OnResponse returned error: %v", err)
	}

	body := []byte(`{"messages":[{"role":"user","content":"what is next?"},{"role":"assistant","content":"the visible answer"}]}`)
	otherCallerReq := &tmodel.InternalLLMRequest{
		Model:               "deepseek-reasoner",
		ReasoningCacheScope: "api_key:2",
		Messages: []tmodel.Message{
			{Role: "user", Content: tmodel.MessageContent{Content: &userContent}},
			{Role: "assistant", Content: tmodel.MessageContent{Content: &assistantContent}},
		},
	}
	modified, err := p.PrepareRequestBody(context.Background(), otherCallerReq, body)
	if err != nil {
		t.Fatalf("PrepareRequestBody returned error: %v", err)
	}
	if strings.Contains(string(modified), "reasoning_content") {
		t.Fatalf("reasoning_content leaked across caller scope: %s", string(modified))
	}

	otherConversationContent := "different prior user message"
	otherConversationBody := []byte(`{"messages":[{"role":"user","content":"different prior user message"},{"role":"assistant","content":"the visible answer"}]}`)
	otherConversationReq := &tmodel.InternalLLMRequest{
		Model:               "deepseek-reasoner",
		ReasoningCacheScope: "api_key:1",
		Messages: []tmodel.Message{
			{Role: "user", Content: tmodel.MessageContent{Content: &otherConversationContent}},
			{Role: "assistant", Content: tmodel.MessageContent{Content: &assistantContent}},
		},
	}
	modified, err = p.PrepareRequestBody(context.Background(), otherConversationReq, otherConversationBody)
	if err != nil {
		t.Fatalf("PrepareRequestBody returned error for other conversation: %v", err)
	}
	if strings.Contains(string(modified), "reasoning_content") {
		t.Fatalf("reasoning_content leaked across conversation scope: %s", string(modified))
	}

	sameConversationReq := &tmodel.InternalLLMRequest{
		Model:               "deepseek-reasoner",
		ReasoningCacheScope: "api_key:1",
		Messages: []tmodel.Message{
			{Role: "user", Content: tmodel.MessageContent{Content: &userContent}},
			{Role: "assistant", Content: tmodel.MessageContent{Content: &assistantContent}},
		},
	}
	modified, err = p.PrepareRequestBody(context.Background(), sameConversationReq, body)
	if err != nil {
		t.Fatalf("PrepareRequestBody returned error for same conversation: %v", err)
	}
	if !strings.Contains(string(modified), `"reasoning_content":"private reasoning"`) {
		t.Fatalf("expected reasoning_content injection for same scope and conversation, got: %s", string(modified))
	}
}

func TestThinkingCacheOnlyAppliesToDeepSeekModels(t *testing.T) {
	p := &DeepSeekThinkingPlugin{
		store: &reasoningStore{
			data: make(map[string]*reasoningEntry),
		},
	}

	userContent := "what is next?"
	assistantContent := "the visible answer"
	reasoningContent := "private reasoning"
	req := &tmodel.InternalLLMRequest{
		Model:               "gpt-4.1",
		ReasoningCacheScope: "api_key:1",
		Messages: []tmodel.Message{{
			Role:    "user",
			Content: tmodel.MessageContent{Content: &userContent},
		}},
	}
	resp := &tmodel.InternalLLMResponse{
		Model: "gpt-4.1",
		Choices: []tmodel.Choice{{
			Message: &tmodel.Message{
				Role:             "assistant",
				Content:          tmodel.MessageContent{Content: &assistantContent},
				ReasoningContent: &reasoningContent,
			},
		}},
	}
	if err := p.OnResponse(context.Background(), req, resp); err != nil {
		t.Fatalf("OnResponse returned error: %v", err)
	}
	if len(p.store.data) != 0 {
		t.Fatalf("expected no cache entry for non-deepseek model, got %d", len(p.store.data))
	}

	body := []byte(`{"messages":[{"role":"user","content":"what is next?"},{"role":"assistant","content":"the visible answer"}]}`)
	injectReq := &tmodel.InternalLLMRequest{
		Model:               "gpt-4.1",
		ReasoningCacheScope: "api_key:1",
		Messages: []tmodel.Message{
			{Role: "user", Content: tmodel.MessageContent{Content: &userContent}},
			{Role: "assistant", Content: tmodel.MessageContent{Content: &assistantContent}},
		},
	}
	modified, err := p.PrepareRequestBody(context.Background(), injectReq, body)
	if err != nil {
		t.Fatalf("PrepareRequestBody returned error: %v", err)
	}
	if string(modified) != string(body) {
		t.Fatalf("expected body to remain unchanged for non-deepseek model, got: %s", string(modified))
	}
}
