package paper

import (
	"testing"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

func TestNormalizeAPIOrigin(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"https://api.anthropic.com", "https://api.anthropic.com"},
		{"https://api.anthropic.com/", "https://api.anthropic.com"},
		{"https://generativelanguage.googleapis.com/v1beta/openai", "https://generativelanguage.googleapis.com"},
		{"https://api.openai.com/v1", "https://api.openai.com"},
	}
	for _, c := range cases {
		if got := normalizeAPIOrigin(c.in); got != c.want {
			t.Fatalf("normalizeAPIOrigin(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLLMRequestURLByProvider(t *testing.T) {
	m := &entity.PaperLLMModelConfig{
		APIBaseURL:   "https://api.anthropic.com",
		ProviderCode: "anthropic",
	}
	if got := llmRequestURL(m); got != "https://api.anthropic.com/v1/messages" {
		t.Fatalf("anthropic: got %q", got)
	}
	m = &entity.PaperLLMModelConfig{
		APIBaseURL:   "https://generativelanguage.googleapis.com/v1beta/openai",
		ProviderCode: "google",
		ModelName:    "gemini-3.8-flash",
	}
	if got := llmRequestURL(m); got != "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions" {
		t.Fatalf("google: got %q", got)
	}
	m = &entity.PaperLLMModelConfig{
		APIBaseURL:   "https://api.openai.com",
		ProviderCode: "openai_compatible",
	}
	if got := llmRequestURL(m); got != "https://api.openai.com/v1/chat/completions" {
		t.Fatalf("openai_compatible: got %q", got)
	}
}
