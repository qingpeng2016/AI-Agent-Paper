package paper

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	httpentity "github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
	httprepo "github.com/qingpeng2016/ai-agent-paper/domain/http/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type LLMUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type LLMChatService struct {
	anthropic httprepo.AnthropicRepo
	gemini    httprepo.GoogleGeminiRepo
	openai    httprepo.OpenAIChatRepo
}

func NewLLMChatService(
	anthropic httprepo.AnthropicRepo,
	gemini httprepo.GoogleGeminiRepo,
	openai httprepo.OpenAIChatRepo,
) *LLMChatService {
	return &LLMChatService{
		anthropic: anthropic,
		gemini:    gemini,
		openai:    openai,
	}
}

func (s *LLMChatService) Complete(ctx context.Context, model *entity.PaperLLMModelConfig, systemPrompt, userPrompt string) (text string, usage LLMUsage, err error) {
	if model == nil {
		return "", usage, errorx.ErrLLMNotConfigured
	}
	key := effectiveLLMAPIKey(model)
	if key == "" {
		return "", usage, errorx.ErrLLMNotConfigured.WithDetail(
			"请在 paper_llm_model_config 填写 api_key，或设置环境变量 ANTHROPIC_API_KEY / OPENAI_API_KEY / GOOGLE_API_KEY",
		)
	}
	timeout := time.Duration(model.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	url := llmRequestURL(model)
	provider := strings.ToLower(strings.TrimSpace(model.ProviderCode))
	switch provider {
	case "anthropic":
		res, callErr := s.anthropic.CreateMessage(ctx, httpentity.AnthropicMessageRequest{
			URL:       url,
			APIKey:    key,
			Model:     model.ModelName,
			System:    systemPrompt,
			User:      userPrompt,
			MaxTokens: 8192,
		})
		if callErr != nil {
			return "", usage, errorx.ErrLLMCallFailed.WithDetail(callErr.Error())
		}
		return res.Text, LLMUsage{PromptTokens: res.InputTokens, CompletionTokens: res.OutputTokens}, nil
	case "google", "gemini":
		res, callErr := s.gemini.ChatCompletions(ctx, httpentity.OpenAIChatCompletionRequest{
			URL:       url,
			APIKey:    key,
			Model:     model.ModelName,
			System:    systemPrompt,
			User:      userPrompt,
			MaxTokens: 8192,
		})
		if callErr != nil {
			return "", usage, errorx.ErrLLMCallFailed.WithDetail(callErr.Error())
		}
		return res.Text, LLMUsage{PromptTokens: res.PromptTokens, CompletionTokens: res.CompletionTokens}, nil
	default:
		res, callErr := s.openai.ChatCompletions(ctx, httpentity.OpenAIChatCompletionRequest{
			URL:       url,
			APIKey:    key,
			Model:     model.ModelName,
			System:    systemPrompt,
			User:      userPrompt,
			MaxTokens: 8192,
		})
		if callErr != nil {
			return "", usage, errorx.ErrLLMCallFailed.WithDetail(callErr.Error())
		}
		return res.Text, LLMUsage{PromptTokens: res.PromptTokens, CompletionTokens: res.CompletionTokens}, nil
	}
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// effectiveLLMAPIKey：库内 key 优先；占位/空时读环境变量（本地开发）。
func effectiveLLMAPIKey(model *entity.PaperLLMModelConfig) string {
	key := strings.TrimSpace(derefString(model.APIKey))
	if key != "" && !strings.Contains(key, "PASTE_YOUR") {
		return key
	}
	provider := strings.ToLower(strings.TrimSpace(model.ProviderCode))
	switch provider {
	case "anthropic":
		if k := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")); k != "" {
			return k
		}
	case "google", "gemini":
		if k := strings.TrimSpace(os.Getenv("GOOGLE_API_KEY")); k != "" {
			return k
		}
		if k := strings.TrimSpace(os.Getenv("GEMINI_API_KEY")); k != "" {
			return k
		}
	case "openai", "openai_compatible", "azure_openai", "gateway", "other":
		if k := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); k != "" {
			return k
		}
	}
	if k := strings.TrimSpace(os.Getenv("PAPER_LLM_API_KEY")); k != "" {
		return k
	}
	return ""
}

func renderPromptTemplate(tmpl string, vars map[string]string) string {
	out := tmpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}
