package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	httpentity "github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
	httprepo "github.com/qingpeng2016/ai-agent-paper/domain/http/repository"
	httpx "github.com/qingpeng2016/ai-agent-paper/infrastructure/http"
)

// Client 实现 domain/http/repository.AnthropicRepo
type Client struct {
	http *httpx.Client
}

func NewClient(http *httpx.Client) httprepo.AnthropicRepo {
	return &Client{http: http}
}

func (c *Client) CreateMessage(ctx context.Context, req httpentity.AnthropicMessageRequest) (httpentity.AnthropicMessageResult, error) {
	var zero httpentity.AnthropicMessageResult
	maxTok := req.MaxTokens
	if maxTok <= 0 {
		maxTok = 8192
	}
	body := map[string]any{
		"model":      req.Model,
		"max_tokens": maxTok,
		"system":     req.System,
		"messages": []map[string]string{
			{"role": "user", "content": req.User},
		},
	}
	headers := map[string]string{
		"x-api-key":         req.APIKey,
		"anthropic-version": "2023-06-01",
	}
	resp, err := c.http.PostJSON(ctx, req.URL, body, headers)
	if err != nil {
		return zero, err
	}
	if resp.StatusCode() >= 400 {
		return zero, httpError(resp.StatusCode(), resp.Body())
	}
	var parsed struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(resp.Body(), &parsed); err != nil {
		return zero, err
	}
	var b strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" {
			b.WriteString(block.Text)
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return zero, fmt.Errorf("empty anthropic content")
	}
	return httpentity.AnthropicMessageResult{
		Text:         out,
		InputTokens:  parsed.Usage.InputTokens,
		OutputTokens: parsed.Usage.OutputTokens,
	}, nil
}

func httpError(status int, body []byte) error {
	const max = 400
	s := strings.TrimSpace(string(body))
	if s == "" {
		return fmt.Errorf("http %d", status)
	}
	if len(s) > max {
		s = s[:max] + "…"
	}
	return fmt.Errorf("http %d: %s", status, s)
}
