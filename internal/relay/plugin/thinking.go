package plugin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	appmodel "github.com/bestruirui/octopus/internal/model"
	"github.com/bestruirui/octopus/internal/op"
	tmodel "github.com/bestruirui/octopus/internal/transformer/model"
	"github.com/bestruirui/octopus/internal/utils/log"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func init() {
	Register(&DeepSeekThinkingPlugin{
		store: &reasoningStore{
			data: make(map[string]*reasoningEntry),
		},
	})
}

type reasoningEntry struct {
	content   string
	timestamp time.Time
}

type reasoningStore struct {
	mu   sync.RWMutex
	data map[string]*reasoningEntry
}

func (s *reasoningStore) Set(key, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = &reasoningEntry{content: content, timestamp: time.Now()}
}

func (s *reasoningStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.data[key]
	if !ok {
		return "", false
	}
	return entry.content, true
}

func (s *reasoningStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *reasoningStore) Cleanup(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for key, entry := range s.data {
		if now.Sub(entry.timestamp) > maxAge {
			delete(s.data, key)
		}
	}
}

type DeepSeekThinkingPlugin struct {
	store    *reasoningStore
	cleanupOnce sync.Once
}

func (p *DeepSeekThinkingPlugin) Name() string {
	return "deepseek_thinking"
}

func (p *DeepSeekThinkingPlugin) enabled() bool {
	enabled, err := op.SettingGetBool(appmodel.SettingKeyEnableThinkingPlugin)
	if err != nil {
		return true
	}
	return enabled
}

func (p *DeepSeekThinkingPlugin) OnRequest(ctx context.Context, req *tmodel.InternalLLMRequest) error {
	return nil
}

func (p *DeepSeekThinkingPlugin) OnResponse(ctx context.Context, req *tmodel.InternalLLMRequest, resp *tmodel.InternalLLMResponse) error {
	if !p.enabled() || resp == nil {
		return nil
	}

	p.startCleanup()

	actualModel := resp.Model
	if actualModel == "" {
		actualModel = req.Model
	}

	for _, choice := range resp.Choices {
		msg := choice.Message
		if msg == nil {
			msg = choice.Delta
		}
		if msg == nil {
			continue
		}

		rc := msg.GetReasoningContent()
		if rc == "" {
			continue
		}

		key := messageKey(actualModel, msg)
		p.store.Set(key, rc)
	}

	return nil
}

func (p *DeepSeekThinkingPlugin) PrepareRequestBody(ctx context.Context, req *tmodel.InternalLLMRequest, body []byte) ([]byte, error) {
	if !p.enabled() || !gjson.ValidBytes(body) {
		return body, nil
	}

	p.startCleanup()

	messages := gjson.GetBytes(body, "messages")
	if !messages.Exists() || !messages.IsArray() {
		return body, nil
	}

	result := body
	var modified bool

	messages.ForEach(func(key, msg gjson.Result) bool {
		role := msg.Get("role").String()
		if role != "assistant" {
			return true
		}

		content := msg.Get("content").String()
		if content == "" {
			return true
		}

		existingRC := msg.Get("reasoning_content").String()
		if existingRC != "" {
			return true
		}

		storeKey := messageKeyFromContent(req.Model, content)
		saved, ok := p.store.Get(storeKey)
		if !ok {
			return true
		}

		msgPath := fmt.Sprintf("messages.%d.reasoning_content", key.Int())
		var err error
		result, err = sjson.SetBytes(result, msgPath, saved)
		if err != nil {
			log.Warnf("plugin thinking: failed to inject reasoning_content: %v", err)
			return true
		}
		modified = true

		p.store.Delete(storeKey)
		return true
	})

	if modified {
		log.Infof("plugin thinking: injected reasoning_content into %d assistant message(s)", 1)
	}

	return result, nil
}

func (p *DeepSeekThinkingPlugin) startCleanup() {
	p.cleanupOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(10 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				p.store.Cleanup(1 * time.Hour)
			}
		}()
	})
}

func messageKey(model string, msg *tmodel.Message) string {
	h := sha256.New()
	h.Write([]byte(model))
	h.Write([]byte{0})
	if msg.Content.Content != nil {
		h.Write([]byte(*msg.Content.Content))
	}
	for _, mc := range msg.Content.MultipleContent {
		if mc.Text != nil {
			h.Write([]byte(*mc.Text))
		}
	}
	for _, tc := range msg.ToolCalls {
		h.Write([]byte(tc.ID))
		h.Write([]byte(tc.Function.Name))
		h.Write([]byte(tc.Function.Arguments))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func messageKeyFromContent(model string, content string) string {
	h := sha256.New()
	h.Write([]byte(model))
	h.Write([]byte{0})
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}
