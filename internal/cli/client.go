package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ApiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) SetToken(token string) {
	c.Token = token
}

func (c *Client) buildURL(path string) string {
	return c.BaseURL + "/" + strings.TrimLeft(path, "/")
}

func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.buildURL(path), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

func (c *Client) decodeResponse(resp *http.Response, target interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr ApiResponse
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Message != "" {
			if resp.StatusCode == http.StatusUnauthorized {
				return ErrUnauthorized
			}
			if resp.StatusCode == http.StatusForbidden {
				return ErrForbidden
			}
			return &CLIError{
				ExitCode: ExitCodeServerError,
				Message:  apiErr.Message,
			}
		}
		return &CLIError{
			ExitCode: ExitCodeServerError,
			Message:  fmt.Sprintf("HTTP %d: %s", resp.StatusCode, tryReadText(body)),
		}
	}

	var apiResp ApiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		if target != nil {
			return json.Unmarshal(body, target)
		}
		return nil
	}

	if target != nil && apiResp.Data != nil {
		return json.Unmarshal(apiResp.Data, target)
	}
	return nil
}

func tryReadText(body []byte) string {
	var s string
	if json.Unmarshal(body, &s) == nil {
		return s
	}
	var m map[string]interface{}
	if json.Unmarshal(body, &m) == nil {
		if msg, ok := m["message"]; ok {
			return fmt.Sprintf("%v", msg)
		}
	}
	return string(body)
}

func (c *Client) Get(path string, target interface{}) error {
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return c.decodeResponse(resp, target)
}

func (c *Client) Post(path string, bodyObj interface{}, target interface{}) error {
	var body io.Reader
	if bodyObj != nil {
		data, err := json.Marshal(bodyObj)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(data)
	}
	resp, err := c.doRequest(http.MethodPost, path, body)
	if err != nil {
		return err
	}
	return c.decodeResponse(resp, target)
}

func (c *Client) Put(path string, bodyObj interface{}, target interface{}) error {
	var body io.Reader
	if bodyObj != nil {
		data, err := json.Marshal(bodyObj)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(data)
	}
	resp, err := c.doRequest(http.MethodPut, path, body)
	if err != nil {
		return err
	}
	return c.decodeResponse(resp, target)
}

func (c *Client) Delete(path string, target interface{}) error {
	resp, err := c.doRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	return c.decodeResponse(resp, target)
}

func (c *Client) Patch(path string, bodyObj interface{}, target interface{}) error {
	var body io.Reader
	if bodyObj != nil {
		data, err := json.Marshal(bodyObj)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(data)
	}
	resp, err := c.doRequest(http.MethodPatch, path, body)
	if err != nil {
		return err
	}
	return c.decodeResponse(resp, target)
}
