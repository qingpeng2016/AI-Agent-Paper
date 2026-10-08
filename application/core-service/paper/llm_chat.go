package paper

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	httpinfra "github.com/qingpeng2016/ai-agent-paper/infrastructure/http"
)

type LLMUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type LLMChatService struct {
	http *httpinfra.Client
}

func NewLLMChatService(http *httpinfra.Client) *LLMChatService {
	return &LLMChatService{http: http}
}

func (s *LLMChatService) Complete(ctx context.Context, model *entity.PaperLLMModelConfig, systemPrompt, userPrompt string) (text string, usage LLMUsage, err error) {
	if model == nil {
		return "", usage, errorx.ErrLLMNotConfigured
	}
	key := strings.TrimSpace(derefString(model.APIKey))
	if key == "" || strings.Contains(key, "PASTE_YOUR") {
		return "", usage, errorx.ErrLLMNotConfigured
	}
	timeout := time.Duration(model.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	provider := strings.ToLower(strings.TrimSpace(model.ProviderCode))
	switch provider {
	case "anthropic":
		return s.completeAnthropic(ctx, model, key, systemPrompt, userPrompt)
	case "openai", "openai_compatible", "azure_openai", "gateway", "other":
		return s.completeOpenAIChat(ctx, model, key, systemPrompt, userPrompt)
	default:
		return s.completeOpenAIChat(ctx, model, key, systemPrompt, userPrompt)
	}
}

func (s *LLMChatService) completeAnthropic(ctx context.Context, model *entity.PaperLLMModelConfig, apiKey, systemPrompt, userPrompt string) (string, LLMUsage, error) {
	var usage LLMUsage
	base := strings.TrimRight(model.APIBaseURL, "/")
	url := base + "/v1/messages"
	body := map[string]any{
		"model":      model.ModelName,
		"max_tokens": 8192,
		"system":     systemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}
	headers := map[string]string{
		"x-api-key":         apiKey,
		"anthropic-version": "2023-06-01",
	}
	resp, err := s.http.PostJSON(ctx, url, body, headers)
	if err != nil {
		return "", usage, errorx.ErrLLMCallFailed
	}
	if resp.StatusCode() >= 400 {
		return "", usage, errorx.ErrLLMCallFailed
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
		return "", usage, errorx.ErrLLMCallFailed
	}
	var b strings.Builder
	for _, c := range parsed.Content {
		if c.Type == "text" {
			b.WriteString(c.Text)
		}
	}
	usage.PromptTokens = parsed.Usage.InputTokens
	usage.CompletionTokens = parsed.Usage.OutputTokens
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "", usage, errorx.ErrLLMCallFailed
	}
	return out, usage, nil
}

func (s *LLMChatService) completeOpenAIChat(ctx context.Context, model *entity.PaperLLMModelConfig, apiKey, systemPrompt, userPrompt string) (string, LLMUsage, error) {
	var usage LLMUsage
	base := strings.TrimRight(model.APIBaseURL, "/")
	url := base + "/v1/chat/completions"
	msgs := []map[string]string{}
	if strings.TrimSpace(systemPrompt) != "" {
		msgs = append(msgs, map[string]string{"role": "system", "content": systemPrompt})
	}
	msgs = append(msgs, map[string]string{"role": "user", "content": userPrompt})
	body := map[string]any{
		"model":    model.ModelName,
		"messages": msgs,
	}
	headers := map[string]string{
		"Authorization": "Bearer " + apiKey,
	}
	resp, err := s.http.PostJSON(ctx, url, body, headers)
	if err != nil {
		return "", usage, errorx.ErrLLMCallFailed
	}
	if resp.StatusCode() >= 400 {
		return "", usage, errorx.ErrLLMCallFailed
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(resp.Body(), &parsed); err != nil {
		return "", usage, errorx.ErrLLMCallFailed
	}
	if len(parsed.Choices) == 0 {
		return "", usage, errorx.ErrLLMCallFailed
	}
	out := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if out == "" {
		return "", usage, errorx.ErrLLMCallFailed
	}
	usage.PromptTokens = parsed.Usage.PromptTokens
	usage.CompletionTokens = parsed.Usage.CompletionTokens
	return out, usage, nil
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func renderPromptTemplate(tmpl string, vars map[string]string) string {
	out := tmpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}
