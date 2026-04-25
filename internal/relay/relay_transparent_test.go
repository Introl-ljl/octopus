package relay

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbmodel "github.com/bestruirui/octopus/internal/model"
	"github.com/bestruirui/octopus/internal/transformer"
	transformerModel "github.com/bestruirui/octopus/internal/transformer/model"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type transparentStubInbound struct{}

func (transparentStubInbound) TransformRequest(ctx context.Context, body []byte) (*transformerModel.InternalLLMRequest, error) {
	return nil, nil
}

func (transparentStubInbound) TransformResponse(ctx context.Context, response *transformerModel.InternalLLMResponse) ([]byte, error) {
	return nil, nil
}

func (transparentStubInbound) TransformStream(ctx context.Context, stream *transformerModel.InternalLLMResponse) ([]byte, error) {
	return nil, nil
}

func (transparentStubInbound) GetInternalResponse(ctx context.Context) (*transformerModel.InternalLLMResponse, error) {
	return nil, nil
}

type transparentStubOutbound struct{}

func (transparentStubOutbound) TransformRequest(ctx context.Context, request *transformerModel.InternalLLMRequest, baseURL, key string) (*http.Request, error) {
	return nil, nil
}

func (transparentStubOutbound) TransformResponse(ctx context.Context, response *http.Response) (*transformerModel.InternalLLMResponse, error) {
	return nil, nil
}

func (transparentStubOutbound) TransformStream(ctx context.Context, eventData []byte) (*transformerModel.InternalLLMResponse, error) {
	return nil, nil
}

func TestBuildTransparentRequestURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		baseURL string
		path    string
		query   string
		want    string
	}{
		{
			name:    "strip duplicated v1 for standard base url",
			baseURL: "https://api.openai.com/v1",
			path:    "/v1/chat/completions",
			query:   "stream=true",
			want:    "https://api.openai.com/v1/chat/completions?stream=true",
		},
		{
			name:    "keep router prefix when base url has no v1",
			baseURL: "https://example.com/proxy",
			path:    "/v1/messages",
			query:   "",
			want:    "https://example.com/proxy/v1/messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, tt.path, http.NoBody)
			c.Request.URL.RawQuery = tt.query

			ra := &relayAttempt{
				relayRequest: &relayRequest{c: c},
				channel:      &dbmodel.Channel{BaseUrls: []dbmodel.BaseUrl{{URL: tt.baseURL}}},
			}

			got, err := ra.buildTransparentRequestURL()
			if err != nil {
				t.Fatalf("buildTransparentRequestURL() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("buildTransparentRequestURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrepareTransparentBodyInjectsStreamUsage(t *testing.T) {
	stream := true
	body := []byte(`{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}],"stream":true,"stream_options":{"include_usage":false}}`)

	ra := &relayAttempt{
		relayRequest: &relayRequest{
			rawBody: body,
			internalRequest: &transformerModel.InternalLLMRequest{
				Model:  "gpt-4.1",
				Stream: &stream,
			},
		},
		outFormat: transformer.FormatOpenAIChat,
	}

	got, err := ra.prepareTransparentBody()
	if err != nil {
		t.Fatalf("prepareTransparentBody() error = %v", err)
	}
	if model := gjson.GetBytes(got, "model").String(); model != "gpt-4.1" {
		t.Fatalf("model = %q, want %q", model, "gpt-4.1")
	}
	if !gjson.GetBytes(got, "stream_options.include_usage").Bool() {
		t.Fatalf("stream_options.include_usage was not forced to true: %s", string(got))
	}
}

func TestHandleTransparentStreamResponsePreservesEventTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stream := true
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", http.NoBody)

	ra := &relayAttempt{
		relayRequest: &relayRequest{
			c:               c,
			inAdapter:       transparentStubInbound{},
			internalRequest: &transformerModel.InternalLLMRequest{Stream: &stream},
			metrics:         &RelayMetrics{},
		},
		outAdapter: transparentStubOutbound{},
	}

	response := &http.Response{
		Header: http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(
			"event: message_start\n" +
				"data: {\"type\":\"message_start\"}\n\n" +
				"event: response.completed\n" +
				"data: {\"type\":\"response.completed\"}\n\n")),
	}

	if err := ra.handleTransparentStreamResponse(context.Background(), response); err != nil {
		t.Fatalf("handleTransparentStreamResponse() error = %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, "event: message_start\n") {
		t.Fatalf("missing message_start event in %q", body)
	}
	if !strings.Contains(body, "event: response.completed\n") {
		t.Fatalf("missing response.completed event in %q", body)
	}
	if strings.Count(body, "event: ") != 2 {
		t.Fatalf("unexpected event frame count in %q", body)
	}
}
