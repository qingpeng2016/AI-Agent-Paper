package openaichat

import (
	"context"
	"strings"

	httpentity "github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
	httpx "github.com/qingpeng2016/ai-agent-paper/infrastructure/http"
)

// DoChatCompletions 共享 OpenAI 兼容 Chat Completions 请求（OpenAI / Gemini 等）。
func DoChatCompletions(ctx context.Context, http *httpx.Client, req httpentity.OpenAIChatCompletionRequest) (httpentity.OpenAIChatCompletionResult, error) {
	var zero httpentity.OpenAIChatCompletionResult
	msgs := []map[string]string{}
	if s := strings.TrimSpace(req.System); s != "" {
		msgs = append(msgs, map[string]string{"role": "system", "content": s})
	}
	msgs = append(msgs, map[string]string{"role": "user", "content": req.User})
	maxTok := req.MaxTokens
	if maxTok <= 0 {
		maxTok = 8192
	}
	body := map[string]any{
		"model":      req.Model,
		"max_tokens": maxTok,
		"messages":   msgs,
	}
	headers := map[string]string{
		"Authorization": "Bearer " + req.APIKey,
	}
	resp, err := http.PostJSON(ctx, req.URL, body, headers)
	if err != nil {
		return zero, err
	}
	if resp.StatusCode() >= 400 {
		return zero, httpError(resp.StatusCode(), resp.Body())
	}
	text, pt, ct, err := parseChatCompletion(resp.Body())
	if err != nil {
		return zero, err
	}
	return httpentity.OpenAIChatCompletionResult{
		Text:             text,
		PromptTokens:     pt,
		CompletionTokens: ct,
	}, nil
}
