package paper

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

// apiPathByProvider 相对 origin 的 HTTP 路径（统一由 provider_code 决定，api_base_url 只存 scheme://host）。
var apiPathByProvider = map[string]string{
	"anthropic":         "/v1/messages",
	"openai":            "/v1/chat/completions",
	"openai_compatible": "/v1/chat/completions",
	"azure_openai":      "/v1/chat/completions",
	"gateway":           "/v1/chat/completions",
	"other":             "/v1/chat/completions",
	"google":            "/v1beta/openai/chat/completions",
	"gemini":            "/v1beta/openai/chat/completions",
}

func normalizeAPIOrigin(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return strings.TrimRight(raw, "/")
	}
	return u.Scheme + "://" + u.Host
}

func apiPathForProvider(model *entity.PaperLLMModelConfig) string {
	if model == nil {
		return "/v1/chat/completions"
	}
	if p := apiPathFromExtra(model.Extra); p != "" {
		return p
	}
	code := strings.ToLower(strings.TrimSpace(model.ProviderCode))
	if p, ok := apiPathByProvider[code]; ok {
		return p
	}
	return "/v1/chat/completions"
}

// apiPathFromExtra 可选覆盖；extra.api_path 或旧字段 chat_completions_path。
func apiPathFromExtra(extra []byte) string {
	if len(extra) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(extra, &m); err != nil {
		return ""
	}
	for _, key := range []string{"api_path", "chat_completions_path"} {
		raw, ok := m[key].(string)
		if !ok {
			continue
		}
		path := strings.TrimSpace(raw)
		if path != "" && strings.HasPrefix(path, "/") {
			return path
		}
	}
	return ""
}

func llmRequestURL(model *entity.PaperLLMModelConfig) string {
	return normalizeAPIOrigin(model.APIBaseURL) + apiPathForProvider(model)
}
