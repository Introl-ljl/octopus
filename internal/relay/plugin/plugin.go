package plugin

import (
	"context"

	"github.com/bestruirui/octopus/internal/transformer/model"
)

type Plugin interface {
	Name() string
	OnRequest(ctx context.Context, req *model.InternalLLMRequest) error
	OnResponse(ctx context.Context, req *model.InternalLLMRequest, resp *model.InternalLLMResponse) error
	PrepareRequestBody(ctx context.Context, req *model.InternalLLMRequest, body []byte) ([]byte, error)
}

var plugins []Plugin

func Register(p Plugin) {
	plugins = append(plugins, p)
}

func RunOnRequest(ctx context.Context, req *model.InternalLLMRequest) error {
	for _, p := range plugins {
		if err := p.OnRequest(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

func RunOnResponse(ctx context.Context, req *model.InternalLLMRequest, resp *model.InternalLLMResponse) error {
	for _, p := range plugins {
		if err := p.OnResponse(ctx, req, resp); err != nil {
			return err
		}
	}
	return nil
}

func RunPrepareRequestBody(ctx context.Context, req *model.InternalLLMRequest, body []byte) ([]byte, error) {
	var err error
	for _, p := range plugins {
		body, err = p.PrepareRequestBody(ctx, req, body)
		if err != nil {
			return nil, err
		}
	}
	return body, nil
}
