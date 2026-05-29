package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/bestruirui/octopus/internal/helper"
	dbmodel "github.com/bestruirui/octopus/internal/model"
	"github.com/bestruirui/octopus/internal/op"
	"github.com/bestruirui/octopus/internal/relay/balancer"
	"github.com/bestruirui/octopus/internal/relay/plugin"
	"github.com/bestruirui/octopus/internal/server/resp"
	"github.com/bestruirui/octopus/internal/transformer"
	"github.com/bestruirui/octopus/internal/transformer/model"
	"github.com/bestruirui/octopus/internal/transformer/outbound"
	"github.com/bestruirui/octopus/internal/utils/log"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/tmaxmax/go-sse"
)

func preferredOutboundTypeForFormat(inFormat transformer.Format) *outbound.OutboundType {
	switch inFormat {
	case transformer.FormatOpenAIChat:
		t := outbound.OutboundTypeOpenAIChat
		return &t
	case transformer.FormatOpenAIResponse:
		t := outbound.OutboundTypeOpenAIResponse
		return &t
	case transformer.FormatOpenAIEmbedding:
		t := outbound.OutboundTypeOpenAIEmbedding
		return &t
	case transformer.FormatAnthropic:
		t := outbound.OutboundTypeAnthropic
		return &t
	default:
		return nil
	}
}

func (ra *relayAttempt) resolvedBaseURL() string {
	if strings.TrimSpace(ra.selectedBaseURL) != "" {
		return ra.selectedBaseURL
	}
	if ra.channel == nil {
		return ""
	}
	return ra.channel.GetBaseUrl()
}

// Handler 处理入站请求并转发到上游服务
func Handler(inFormat transformer.Format, c *gin.Context) {
	// 解析请求
	internalRequest, rawBody, inAdapter, err := parseRequest(inFormat, c)
	if err != nil {
		return
	}
	ctx := c.Request.Context()
	apiKeyID := c.GetInt("api_key_id")
	internalRequest.ReasoningCacheScope = fmt.Sprintf("api_key:%d", apiKeyID)

	if err := plugin.RunOnRequest(ctx, internalRequest); err != nil {
		log.Warnf("plugin OnRequest failed: %v", err)
	}
	supportedModels := c.GetString("supported_models")
	if supportedModels != "" {
		supportedModelsArray := strings.Split(supportedModels, ",")
		if !slices.Contains(supportedModelsArray, internalRequest.Model) {
			resp.Error(c, http.StatusBadRequest, "model not supported")
			return
		}
	}

	requestModel := internalRequest.Model

	// 获取通道分组
	group, err := op.GroupGetEnabledMap(requestModel, c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusNotFound, "model not found")
		return
	}

	// 创建迭代器（策略排序 + 粘性优先）
	iter := balancer.NewIterator(group, apiKeyID, requestModel)
	if iter.Len() == 0 {
		resp.Error(c, http.StatusServiceUnavailable, "no available channel")
		return
	}

	// 初始化 Metrics
	metrics := NewRelayMetrics(apiKeyID, requestModel, string(inFormat), internalRequest)

	// 请求级上下文
	req := &relayRequest{
		c:               c,
		inFormat:        inFormat,
		rawBody:         rawBody,
		inAdapter:       inAdapter,
		internalRequest: internalRequest,
		metrics:         metrics,
		apiKeyID:        apiKeyID,
		requestModel:    requestModel,
		iter:            iter,
	}

	var lastErr error

	for iter.Next() {
		select {
		case <-c.Request.Context().Done():
			log.Infof("request context canceled, stopping retry")
			metrics.Save(c.Request.Context(), false, context.Canceled, iter.Attempts())
			return
		default:
		}

		item := iter.Item()

		// 获取通道
		channel, err := op.ChannelGet(item.ChannelID, c.Request.Context())
		if err != nil {
			log.Warnf("failed to get channel %d: %v", item.ChannelID, err)
			iter.Skip(item.ChannelID, 0, fmt.Sprintf("channel_%d", item.ChannelID), fmt.Sprintf("channel not found: %v", err))
			lastErr = err
			continue
		}
		if !channel.Enabled {
			iter.Skip(channel.ID, 0, channel.Name, "channel disabled")
			continue
		}

		usedKey := channel.GetChannelKey()
		if usedKey.ChannelKey == "" {
			iter.Skip(channel.ID, 0, channel.Name, "no available key")
			continue
		}

		// 熔断检查
		if iter.SkipCircuitBreak(channel.ID, usedKey.ID, channel.Name) {
			continue
		}

		selectedBaseURL, selectedType := channel.GetBaseUrlByType(preferredOutboundTypeForFormat(inFormat))
		if selectedBaseURL == "" {
			iter.Skip(channel.ID, usedKey.ID, channel.Name, "no available base url")
			continue
		}

		// 出站适配器
		outAdapter := outbound.Get(selectedType)
		if outAdapter == nil {
			iter.Skip(channel.ID, usedKey.ID, channel.Name, fmt.Sprintf("unsupported channel type: %d", selectedType))
			continue
		}

		// 类型兼容性检查
		if internalRequest.IsEmbeddingRequest() && !outbound.IsEmbeddingChannelType(selectedType) {
			iter.Skip(channel.ID, usedKey.ID, channel.Name, "channel type not compatible with embedding request")
			continue
		}
		if internalRequest.IsChatRequest() && !outbound.IsChatChannelType(selectedType) {
			iter.Skip(channel.ID, usedKey.ID, channel.Name, "channel type not compatible with chat request")
			continue
		}

		// 设置实际模型
		internalRequest.Model = item.ModelName

		log.Infof("request model %s, mode: %d, forwarding to channel: %s model: %s (attempt %d/%d, sticky=%t)",
			requestModel, group.Mode, channel.Name, item.ModelName,
			iter.Index()+1, iter.Len(), iter.IsSticky())

		// 出站格式
		outFormat := transformer.Format(outbound.OutboundTypeToFormat(selectedType))
		metrics.OutboundFormat = string(outFormat)

		// 构造尝试级上下文 -- 只写变化的字段
		ra := &relayAttempt{
			relayRequest:         req,
			outFormat:            outFormat,
			outAdapter:           outAdapter,
			channel:              channel,
			selectedBaseURL:      selectedBaseURL,
			selectedType:         selectedType,
			usedKey:              usedKey,
			firstTokenTimeOutSec: group.FirstTokenTimeOut,
		}

		result := ra.attempt()
		if result.Success {
			if ra.metrics.InternalResponse != nil {
				if err := plugin.RunOnResponse(ctx, internalRequest, ra.metrics.InternalResponse); err != nil {
					log.Warnf("plugin OnResponse failed: %v", err)
				}
			}
			metrics.Save(c.Request.Context(), true, nil, iter.Attempts())
			return
		}
		if result.Written {
			metrics.Save(c.Request.Context(), false, result.Err, iter.Attempts())
			return
		}
		lastErr = result.Err
	}

	// 所有通道都失败
	metrics.Save(c.Request.Context(), false, lastErr, iter.Attempts())
	resp.Error(c, http.StatusBadGateway, "all channels failed")
}

// attempt 统一管理一次通道尝试的完整生命周期
func (ra *relayAttempt) attempt() attemptResult {
	span := ra.iter.StartAttempt(ra.channel.ID, ra.usedKey.ID, ra.channel.Name)

	// 转发请求
	statusCode, fwdErr := ra.forward()

	// 更新 channel key 状态
	ra.usedKey.StatusCode = statusCode
	ra.usedKey.LastUseTimeStamp = time.Now().Unix()

	if fwdErr == nil {
		// ====== 成功 ======
		ra.collectResponse()
		ra.usedKey.TotalCost += ra.metrics.Stats.InputCost + ra.metrics.Stats.OutputCost
		op.ChannelKeyUpdate(ra.usedKey)

		span.End(dbmodel.AttemptSuccess, statusCode, "")

		// Channel 维度统计
		op.StatsChannelUpdate(ra.channel.ID, dbmodel.StatsMetrics{
			WaitTime:       span.Duration().Milliseconds(),
			RequestSuccess: 1,
		})

		// 熔断器：记录成功
		balancer.RecordSuccess(ra.channel.ID, ra.usedKey.ID, ra.internalRequest.Model)
		// 会话保持：更新粘性记录
		balancer.SetSticky(ra.apiKeyID, ra.requestModel, ra.channel.ID, ra.usedKey.ID)

		ra.metrics.ParamOverride = paramOverrideValue(ra.channel.ParamOverride)

		return attemptResult{Success: true}
	}

	// ====== 失败 ======
	op.ChannelKeyUpdate(ra.usedKey)
	span.End(dbmodel.AttemptFailed, statusCode, fwdErr.Error())

	// Channel 维度统计
	op.StatsChannelUpdate(ra.channel.ID, dbmodel.StatsMetrics{
		WaitTime:      span.Duration().Milliseconds(),
		RequestFailed: 1,
	})

	// 熔断器：记录失败
	balancer.RecordFailure(ra.channel.ID, ra.usedKey.ID, ra.internalRequest.Model)

	ra.metrics.ParamOverride = paramOverrideValue(ra.channel.ParamOverride)

	written := ra.c.Writer.Written()
	if written {
		ra.collectResponse()
	}
	return attemptResult{
		Success: false,
		Written: written,
		Err:     fmt.Errorf("channel %s failed: %v", ra.channel.Name, fwdErr),
	}
}

// parseRequest 解析并验证入站请求
func parseRequest(inFormat transformer.Format, c *gin.Context) (*model.InternalLLMRequest, []byte, model.Inbound, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return nil, nil, nil, err
	}

	inAdapter := transformer.Default().GetInbound(inFormat)
	if inAdapter == nil {
		resp.Error(c, http.StatusInternalServerError, "unsupported inbound format")
		return nil, nil, nil, fmt.Errorf("unsupported inbound format: %s", inFormat)
	}

	internalRequest, err := inAdapter.TransformRequest(c.Request.Context(), body)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return nil, nil, nil, err
	}

	// Pass through the original query parameters
	internalRequest.Query = c.Request.URL.Query()

	if err := internalRequest.Validate(); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return nil, nil, nil, err
	}

	return internalRequest, body, inAdapter, nil
}

// forward 转发请求到上游服务
func (ra *relayAttempt) forward() (int, error) {
	ctx := ra.c.Request.Context()

	// 判断是否可以走透明代理
	if !transformer.Default().NeedTransform(ra.inFormat, ra.outFormat) {
		log.Infof("transparent proxy: %s -> %s", ra.inFormat, ra.outFormat)
		ra.metrics.IsDirect = true
		return ra.forwardTransparent(ctx)
	}

	// ==================== 原有转换逻辑 ====================

	// 构建出站请求
	outboundRequest, err := ra.outAdapter.TransformRequest(
		ctx,
		ra.internalRequest,
		ra.resolvedBaseURL(),
		ra.usedKey.ChannelKey,
	)
	if err != nil {
		log.Warnf("failed to create request: %v", err)
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// 应用 ParamOverride 到请求体
	if ra.channel.ParamOverride != nil && *ra.channel.ParamOverride != "" {
		body, err := io.ReadAll(outboundRequest.Body)
		if err != nil {
			return 0, fmt.Errorf("failed to read body: %w", err)
		}

		var bodyMap map[string]any
		if err := json.Unmarshal(body, &bodyMap); err != nil {
			log.Warnf("failed to unmarshal request body: %v, skipping param_override", err)
			outboundRequest.Body = io.NopCloser(bytes.NewBuffer(body))
			return 0, nil
		}
		var override map[string]any
		if err := json.Unmarshal([]byte(*ra.channel.ParamOverride), &override); err != nil {
			log.Warnf("failed to unmarshal param_override: %v, skipping", err)
			outboundRequest.Body = io.NopCloser(bytes.NewBuffer(body))
			return 0, nil
		}
		maps.Copy(bodyMap, override)
		modifiedBody, err := json.Marshal(bodyMap)
		if err != nil {
			log.Warnf("failed to marshal modified body: %v, skipping param_override", err)
			outboundRequest.Body = io.NopCloser(bytes.NewBuffer(body))
			return 0, nil
		}
		outboundRequest.Body = io.NopCloser(bytes.NewBuffer(modifiedBody))
		outboundRequest.ContentLength = int64(len(modifiedBody))
	}

	// 插件：注入 reasoning_content
	if body, err := io.ReadAll(outboundRequest.Body); err == nil {
		if modified, err := plugin.RunPrepareRequestBody(ctx, ra.internalRequest, body); err == nil && len(modified) > 0 {
			outboundRequest.Body = io.NopCloser(bytes.NewBuffer(modified))
			outboundRequest.ContentLength = int64(len(modified))
		} else {
			outboundRequest.Body = io.NopCloser(bytes.NewBuffer(body))
		}
	}

	// 复制请求头
	ra.copyHeaders(outboundRequest)

	// 发送请求
	response, err := ra.sendRequest(outboundRequest)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	// 检查响应状态
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return 0, fmt.Errorf("failed to read response body: %w", err)
		}
		return 0, fmt.Errorf("upstream error: %d: %s", response.StatusCode, string(body))
	}

	// 处理响应
	if ra.internalRequest.Stream != nil && *ra.internalRequest.Stream {
		if err := ra.handleStreamResponse(ctx, response); err != nil {
			return 0, err
		}
		return response.StatusCode, nil
	}
	if err := ra.handleResponse(ctx, response); err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}

// copyHeaders 复制请求头，过滤 hop-by-hop 头
func (ra *relayAttempt) copyHeaders(outboundRequest *http.Request) {
	for key, values := range ra.c.Request.Header {
		if hopByHopHeaders[strings.ToLower(key)] {
			continue
		}
		for _, value := range values {
			outboundRequest.Header.Set(key, value)
		}
	}
	if len(ra.channel.CustomHeader) > 0 {
		for _, header := range ra.channel.CustomHeader {
			outboundRequest.Header.Set(header.HeaderKey, header.HeaderValue)
		}
	}
}

// sendRequest 发送 HTTP 请求
func (ra *relayAttempt) sendRequest(req *http.Request) (*http.Response, error) {
	httpClient, err := helper.ChannelHttpClient(ra.channel)
	if err != nil {
		log.Warnf("failed to get http client: %v", err)
		return nil, err
	}

	response, err := httpClient.Do(req)
	if err != nil {
		log.Warnf("failed to send request: %v", err)
		return nil, err
	}

	return response, nil
}

// handleStreamResponse 处理流式响应
func (ra *relayAttempt) handleStreamResponse(ctx context.Context, response *http.Response) error {
	if ct := response.Header.Get("Content-Type"); ct != "" && !strings.Contains(strings.ToLower(ct), "text/event-stream") {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 16*1024))
		return fmt.Errorf("upstream returned non-SSE content-type %q for stream request: %s", ct, string(body))
	}

	// 设置 SSE 响应头
	ra.c.Header("Content-Type", "text/event-stream")
	ra.c.Header("Cache-Control", "no-cache")
	ra.c.Header("Connection", "keep-alive")
	ra.c.Header("X-Accel-Buffering", "no")

	firstToken := true

	type sseReadResult struct {
		data string
		err  error
	}
	results := make(chan sseReadResult, 1)
	go func() {
		defer close(results)
		readCfg := &sse.ReadConfig{MaxEventSize: maxSSEEventSize}
		for ev, err := range sse.Read(response.Body, readCfg) {
			if err != nil {
				results <- sseReadResult{err: err}
				return
			}
			results <- sseReadResult{data: ev.Data}
		}
	}()

	var firstTokenTimer *time.Timer
	var firstTokenC <-chan time.Time
	if firstToken && ra.firstTokenTimeOutSec > 0 {
		firstTokenTimer = time.NewTimer(time.Duration(ra.firstTokenTimeOutSec) * time.Second)
		firstTokenC = firstTokenTimer.C
		defer func() {
			if firstTokenTimer != nil {
				firstTokenTimer.Stop()
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			log.Infof("client disconnected, stopping stream")
			return nil
		case <-firstTokenC:
			log.Warnf("first token timeout (%ds), switching channel", ra.firstTokenTimeOutSec)
			_ = response.Body.Close()
			return fmt.Errorf("first token timeout (%ds)", ra.firstTokenTimeOutSec)
		case r, ok := <-results:
			if !ok {
				log.Infof("stream end")
				return nil
			}
			if r.err != nil {
				log.Warnf("failed to read event: %v", r.err)
				return fmt.Errorf("failed to read stream event: %w", r.err)
			}

			data, err := ra.transformStreamData(ctx, r.data)
			if err != nil || len(data) == 0 {
				continue
			}
			if firstToken {
				ra.metrics.SetFirstTokenTime(time.Now())
				firstToken = false
				if firstTokenTimer != nil {
					if !firstTokenTimer.Stop() {
						select {
						case <-firstTokenTimer.C:
						default:
						}
					}
					firstTokenTimer = nil
					firstTokenC = nil
				}
			}

			ra.c.Writer.Write(data)
			ra.c.Writer.Flush()
		}
	}
}

// transformStreamData 转换流式数据
func (ra *relayAttempt) transformStreamData(ctx context.Context, data string) ([]byte, error) {
	internalStream, err := ra.outAdapter.TransformStream(ctx, []byte(data))
	if err != nil {
		log.Warnf("failed to transform stream: %v", err)
		return nil, err
	}
	if internalStream == nil {
		return nil, nil
	}

	inStream, err := ra.inAdapter.TransformStream(ctx, internalStream)
	if err != nil {
		log.Warnf("failed to transform stream: %v", err)
		return nil, err
	}

	return inStream, nil
}

// handleResponse 处理非流式响应
func (ra *relayAttempt) handleResponse(ctx context.Context, response *http.Response) error {
	internalResponse, err := ra.outAdapter.TransformResponse(ctx, response)
	if err != nil {
		log.Warnf("failed to transform response: %v", err)
		return fmt.Errorf("failed to transform outbound response: %w", err)
	}

	inResponse, err := ra.inAdapter.TransformResponse(ctx, internalResponse)
	if err != nil {
		log.Warnf("failed to transform response: %v", err)
		return fmt.Errorf("failed to transform inbound response: %w", err)
	}

	ra.c.Data(http.StatusOK, "application/json", inResponse)
	return nil
}

// collectResponse 收集响应信息
func (ra *relayAttempt) collectResponse() {
	internalResponse, err := ra.inAdapter.GetInternalResponse(ra.c.Request.Context())
	if err != nil || internalResponse == nil {
		return
	}

	ra.metrics.SetInternalResponse(internalResponse, ra.internalRequest.Model)
}

// ============================== 透明代理核心逻辑 ==============================

// forwardTransparent 执行透明代理，跳过所有的内部模型转换，直接转发原始 JSON 请求和响应
func (ra *relayAttempt) forwardTransparent(ctx context.Context) (int, error) {
	// 1. 对透明请求做最小必要改写：替换 model，并为 OpenAI chat stream 保留 usage 统计。
	newBody, err := ra.prepareTransparentBody()
	if err != nil {
		return 0, err
	}

	// 2. 构造 HTTP 请求
	requestURL, err := ra.buildTransparentRequestURL()
	if err != nil {
		return 0, err
	}

	outboundRequest, err := http.NewRequestWithContext(ctx, ra.c.Request.Method, requestURL, bytes.NewReader(newBody))
	if err != nil {
		return 0, fmt.Errorf("failed to create transparent request: %w", err)
	}

	// 设置认证头 (兼容多种协议)
	if ra.outFormat == transformer.FormatAnthropic {
		outboundRequest.Header.Set("x-api-key", ra.usedKey.ChannelKey)
		outboundRequest.Header.Set("anthropic-version", "2023-06-01")
	} else {
		outboundRequest.Header.Set("Authorization", "Bearer "+ra.usedKey.ChannelKey)
	}
	outboundRequest.Header.Set("Content-Type", "application/json")

	// 插件：注入 reasoning_content
	if body, err := io.ReadAll(outboundRequest.Body); err == nil {
		if modified, err := plugin.RunPrepareRequestBody(ctx, ra.internalRequest, body); err == nil && len(modified) > 0 {
			outboundRequest.Body = io.NopCloser(bytes.NewBuffer(modified))
			outboundRequest.ContentLength = int64(len(modified))
		} else {
			outboundRequest.Body = io.NopCloser(bytes.NewBuffer(body))
		}
	}

	ra.copyHeaders(outboundRequest)

	// 3. 发送请求
	response, err := ra.sendRequest(outboundRequest)
	if err != nil {
		return 0, fmt.Errorf("failed to send transparent request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(response.Body)
		return 0, fmt.Errorf("upstream error: %d: %s", response.StatusCode, string(body))
	}

	// 4. 处理响应（透明透传）
	isStream := ra.internalRequest.Stream != nil && *ra.internalRequest.Stream
	if isStream {
		if err := ra.handleTransparentStreamResponse(ctx, response); err != nil {
			return 0, err
		}
	} else {
		if err := ra.handleTransparentResponse(ctx, response); err != nil {
			return 0, err
		}
	}

	return response.StatusCode, nil
}

func (ra *relayAttempt) prepareTransparentBody() ([]byte, error) {
	newBody := ra.rawBody
	if gjson.GetBytes(newBody, "model").Exists() {
		var err error
		newBody, err = sjson.SetBytes(newBody, "model", ra.internalRequest.Model)
		if err != nil {
			log.Warnf("failed to set model in transparent proxy: %v", err)
			return nil, fmt.Errorf("failed to prepare transparent request: %w", err)
		}
	}

	if !ra.shouldInjectTransparentUsage() {
		return newBody, nil
	}

	newBody, err := sjson.SetBytes(newBody, "stream_options.include_usage", true)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare transparent request: %w", err)
	}

	return newBody, nil
}

func (ra *relayAttempt) shouldInjectTransparentUsage() bool {
	return ra.outFormat == transformer.FormatOpenAIChat &&
		ra.internalRequest.Stream != nil && *ra.internalRequest.Stream
}

func (ra *relayAttempt) buildTransparentRequestURL() (string, error) {
	baseURL, err := url.Parse(strings.TrimSuffix(ra.resolvedBaseURL(), "/"))
	if err != nil {
		return "", fmt.Errorf("failed to parse transparent base url: %w", err)
	}

	basePath := strings.TrimSuffix(baseURL.Path, "/")
	requestPath := ra.transparentUpstreamPath(basePath)
	switch {
	case basePath == "":
		baseURL.Path = requestPath
	case requestPath == "":
		baseURL.Path = basePath
	default:
		baseURL.Path = basePath + "/" + strings.TrimPrefix(requestPath, "/")
	}
	baseURL.RawQuery = ra.c.Request.URL.RawQuery

	return baseURL.String(), nil
}

func (ra *relayAttempt) transparentUpstreamPath(basePath string) string {
	requestPath := ra.c.Request.URL.Path
	if requestPath == "" {
		return ""
	}
	if basePath == "/v1" || strings.HasSuffix(basePath, "/v1") {
		if requestPath == "/v1" {
			return ""
		}
		if strings.HasPrefix(requestPath, "/v1/") {
			return strings.TrimPrefix(requestPath, "/v1")
		}
	}
	return requestPath
}

// handleTransparentResponse 直接透传非流式响应
func (ra *relayAttempt) handleTransparentResponse(ctx context.Context, response *http.Response) error {
	for k, v := range response.Header {
		if hopByHopHeaders[strings.ToLower(k)] {
			continue
		}
		for _, vv := range v {
			ra.c.Header(k, vv)
		}
	}
	ra.c.Status(response.StatusCode)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	_, err = ra.c.Writer.Write(body)

	// 为了统计计费，我们需要将原始响应转码一遍，这是必要的开销
	// 如果需要极致性能，可以考虑直接解析 json 提取 usage
	if internalResp, parseErr := ra.outAdapter.TransformResponse(ctx, &http.Response{
		StatusCode: response.StatusCode,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     response.Header,
	}); parseErr == nil {
		ra.metrics.SetInternalResponse(internalResp, ra.internalRequest.Model)
	}

	return err
}

// handleTransparentStreamResponse 直接透传流式响应（SSE）
func (ra *relayAttempt) handleTransparentStreamResponse(ctx context.Context, response *http.Response) error {
	ra.c.Header("Content-Type", "text/event-stream")
	ra.c.Header("Cache-Control", "no-cache")
	ra.c.Header("Connection", "keep-alive")
	ra.c.Header("X-Accel-Buffering", "no")

	for k, v := range response.Header {
		if hopByHopHeaders[strings.ToLower(k)] || strings.ToLower(k) == "content-type" || strings.ToLower(k) == "content-length" {
			continue
		}
		for _, vv := range v {
			ra.c.Header(k, vv)
		}
	}

	firstToken := true
	var firstTokenTimer *time.Timer
	var firstTokenC <-chan time.Time

	if ra.firstTokenTimeOutSec > 0 {
		firstTokenTimer = time.NewTimer(time.Duration(ra.firstTokenTimeOutSec) * time.Second)
		firstTokenC = firstTokenTimer.C
		defer func() {
			if firstTokenTimer != nil {
				firstTokenTimer.Stop()
			}
		}()
	}

	type transparentSSEReadResult struct {
		event sse.Event
		err   error
	}
	results := make(chan transparentSSEReadResult, 1)

	go func() {
		defer close(results)
		readCfg := &sse.ReadConfig{MaxEventSize: maxSSEEventSize}
		for ev, err := range sse.Read(response.Body, readCfg) {
			if err != nil {
				results <- transparentSSEReadResult{err: err}
				return
			}
			results <- transparentSSEReadResult{event: ev}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-firstTokenC:
			return fmt.Errorf("first token timeout (%ds)", ra.firstTokenTimeOutSec)
		case r, ok := <-results:
			if !ok {
				return nil
			}
			if r.err != nil {
				return fmt.Errorf("failed to read stream event: %w", r.err)
			}

			if firstToken {
				ra.metrics.SetFirstTokenTime(time.Now())
				firstToken = false
				if firstTokenTimer != nil {
					firstTokenTimer.Stop()
					firstTokenTimer = nil
					firstTokenC = nil
				}
			}

			if err := writeTransparentSSEMessage(ra.c.Writer, r.event); err != nil {
				return fmt.Errorf("failed to write stream event: %w", err)
			}
			ra.c.Writer.Flush()

			// 后台异步进行流式解析统计 usage
			if internalStream, err := ra.outAdapter.TransformStream(ctx, []byte(r.event.Data)); err == nil && internalStream != nil {
				// 获取 inbound 实例保存状态，以便请求结束时 collectResponse 获取最终结果
				ra.inAdapter.TransformStream(ctx, internalStream)
			}
		}
	}
}

func writeTransparentSSEMessage(w io.Writer, ev sse.Event) error {
	if ev.Type != "" {
		if strings.ContainsAny(ev.Type, "\r\n") {
			return fmt.Errorf("invalid sse event type")
		}
		if _, err := io.WriteString(w, "event: "+ev.Type+"\n"); err != nil {
			return err
		}
	}

	if ev.Data != "" {
		data := strings.ReplaceAll(ev.Data, "\r\n", "\n")
		data = strings.ReplaceAll(data, "\r", "\n")
		for _, line := range strings.Split(data, "\n") {
			if _, err := io.WriteString(w, "data: "+line+"\n"); err != nil {
				return err
			}
		}
	}

	_, err := io.WriteString(w, "\n")
	return err
}

func paramOverrideValue(ptr *string) string {
	if ptr == nil || *ptr == "" {
		return ""
	}
	return *ptr
}
