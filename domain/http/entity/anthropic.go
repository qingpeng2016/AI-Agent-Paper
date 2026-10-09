package entity

// AnthropicMessageRequest Messages API（由调用方按 paper_llm_model_config 拼好完整 URL）。
type AnthropicMessageRequest struct {
	URL       string
	APIKey    string
	Model     string
	System    string
	User      string
	MaxTokens int
}

type AnthropicMessageResult struct {
	Text           string
	InputTokens    int
	OutputTokens   int
}
