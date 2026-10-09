package entity

// OpenAIChatCompletionRequest Chat Completions（OpenAI 兼容：OpenAI 网关 / Google Gemini 等）。
type OpenAIChatCompletionRequest struct {
	URL       string
	APIKey    string
	Model     string
	System    string
	User      string
	MaxTokens int
}

type OpenAIChatCompletionResult struct {
	Text             string
	PromptTokens     int
	CompletionTokens int
}
