package llm

type LLMUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

func (u LLMUsage) TotalTokens() int {
	return u.PromptTokens + u.CompletionTokens
}
