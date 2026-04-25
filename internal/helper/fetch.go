package helper

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bestruirui/octopus/internal/model"
	"github.com/bestruirui/octopus/internal/transformer/outbound"
	"github.com/dlclark/regexp2"
)

func FetchModels(ctx context.Context, request model.Channel) ([]string, error) {
	client, err := ChannelHttpClient(&request)
	if err != nil {
		return nil, err
	}
	fetchModelsByType := func(req model.Channel, resolvedType outbound.OutboundType) ([]string, error) {
		switch resolvedType {
		case outbound.OutboundTypeAnthropic:
			return fetchAnthropicModels(client, ctx, req)
		case outbound.OutboundTypeGemini:
			return fetchGeminiModels(client, ctx, req)
		default:
			return fetchOpenAIModels(client, ctx, req)
		}
	}

	applyMatchRegex := func(models []string) ([]string, error) {
		if request.MatchRegex != nil && *request.MatchRegex != "" {
			matchModel := make([]string, 0)
			re, err := regexp2.Compile(*request.MatchRegex, regexp2.ECMAScript)
			if err != nil {
				return nil, err
			}
			for _, model := range models {
				matched, err := re.MatchString(model)
				if err != nil {
					return nil, err
				}
				if matched {
					matchModel = append(matchModel, model)
				}
			}
			return matchModel, nil
		}
		return models, nil
	}

	if len(request.BaseUrls) == 0 {
		_, resolvedType := request.GetBaseUrlByType(nil)
		fetchModel, err := fetchModelsByType(request, resolvedType)
		if err != nil {
			return nil, err
		}
		return applyMatchRegex(fetchModel)
	}

	mergedModels := make([]string, 0)
	seen := make(map[string]struct{}, len(request.BaseUrls))
	var lastErr error

	for _, baseURL := range request.BaseUrls {
		if strings.TrimSpace(baseURL.URL) == "" {
			continue
		}
		resolvedType := request.Type
		if baseURL.Type != nil {
			resolvedType = *baseURL.Type
		}
		childRequest := request
		childRequest.Type = resolvedType
		childRequest.BaseUrls = []model.BaseUrl{baseURL}

		fetchModel, err := fetchModelsByType(childRequest, resolvedType)
		if err != nil {
			lastErr = err
			continue
		}
		for _, name := range fetchModel {
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			mergedModels = append(mergedModels, name)
		}
	}

	if len(mergedModels) == 0 && lastErr != nil {
		return nil, lastErr
	}

	return applyMatchRegex(mergedModels)
}

// refer: https://platform.openai.com/docs/api-reference/models/list
func fetchOpenAIModels(client *http.Client, ctx context.Context, request model.Channel) ([]string, error) {
	req, _ := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		request.GetBaseUrl()+"/models",
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+request.GetChannelKey().ChannelKey)
	for _, header := range request.CustomHeader {
		if header.HeaderKey != "" {
			req.Header.Set(header.HeaderKey, header.HeaderValue)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result model.OpenAIModelList

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]string, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, m.ID)
	}
	return models, nil
}

// refer: https://ai.google.dev/api/models
func fetchGeminiModels(client *http.Client, ctx context.Context, request model.Channel) ([]string, error) {
	var allModels []string
	pageToken := ""

	for {
		req, _ := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			request.GetBaseUrl()+"/models",
			nil,
		)
		req.Header.Set("X-Goog-Api-Key", request.GetChannelKey().ChannelKey)
		for _, header := range request.CustomHeader {
			if header.HeaderKey != "" {
				req.Header.Set(header.HeaderKey, header.HeaderValue)
			}
		}
		if pageToken != "" {
			q := req.URL.Query()
			q.Add("pageToken", pageToken)
			req.URL.RawQuery = q.Encode()
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result model.GeminiModelList

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		for _, m := range result.Models {
			name := strings.TrimPrefix(m.Name, "models/")
			allModels = append(allModels, name)
		}

		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}
	if len(allModels) == 0 {
		return fetchOpenAIModels(client, ctx, request)
	}
	return allModels, nil
}

// refer: https://platform.claude.com/docs
func fetchAnthropicModels(client *http.Client, ctx context.Context, request model.Channel) ([]string, error) {

	var allModels []string
	var afterID string
	for {

		req, _ := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			request.GetBaseUrl()+"/models",
			nil,
		)
		req.Header.Set("X-Api-Key", request.GetChannelKey().ChannelKey)
		req.Header.Set("Anthropic-Version", "2023-06-01")
		for _, header := range request.CustomHeader {
			if header.HeaderKey != "" {
				req.Header.Set(header.HeaderKey, header.HeaderValue)
			}
		}
		// 设置多页参数
		q := req.URL.Query()

		if afterID != "" {
			q.Set("after_id", afterID)
		}
		req.URL.RawQuery = q.Encode()

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result model.AnthropicModelList

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		for _, m := range result.Data {
			allModels = append(allModels, m.ID)
		}

		if !result.HasMore {
			break
		}

		afterID = result.LastID
	}
	if len(allModels) == 0 {
		return fetchOpenAIModels(client, ctx, request)
	}
	return allModels, nil
}
