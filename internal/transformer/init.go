package transformer

import (
	anthropicInbound "github.com/bestruirui/octopus/internal/transformer/inbound/anthropic"
	openaiInbound "github.com/bestruirui/octopus/internal/transformer/inbound/openai"
	"github.com/bestruirui/octopus/internal/transformer/model"
	anthropicOutbound "github.com/bestruirui/octopus/internal/transformer/outbound/authropic"
	geminiOutbound "github.com/bestruirui/octopus/internal/transformer/outbound/gemini"
	openaiOutbound "github.com/bestruirui/octopus/internal/transformer/outbound/openai"
	volcengineOutbound "github.com/bestruirui/octopus/internal/transformer/outbound/volcengine"
)

func init() {
	r := Default()

	// 注册入站适配器
	r.RegisterInbound(FormatOpenAIChat, func() model.Inbound { return &openaiInbound.ChatInbound{} })
	r.RegisterInbound(FormatOpenAIResponse, func() model.Inbound { return &openaiInbound.ResponseInbound{} })
	r.RegisterInbound(FormatOpenAIEmbedding, func() model.Inbound { return &openaiInbound.EmbeddingInbound{} })
	r.RegisterInbound(FormatAnthropic, func() model.Inbound { return &anthropicInbound.MessagesInbound{} })

	// 注册出站适配器
	r.RegisterOutbound(FormatOpenAIChat, func() model.Outbound { return &openaiOutbound.ChatOutbound{} })
	r.RegisterOutbound(FormatOpenAIResponse, func() model.Outbound { return &openaiOutbound.ResponseOutbound{} })
	r.RegisterOutbound(FormatOpenAIEmbedding, func() model.Outbound { return &openaiOutbound.EmbeddingOutbound{} })
	r.RegisterOutbound(FormatAnthropic, func() model.Outbound { return &anthropicOutbound.MessageOutbound{} })
	r.RegisterOutbound(FormatGemini, func() model.Outbound { return &geminiOutbound.MessagesOutbound{} })
	r.RegisterOutbound(FormatVolcengine, func() model.Outbound { return &volcengineOutbound.ResponseOutbound{} })
}
