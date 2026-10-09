package openaichat

import (
	"encoding/json"
	"fmt"
	"strings"
)

func parseChatCompletion(body []byte) (text string, promptTokens, completionTokens int, err error) {
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
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", 0, 0, err
	}
	if len(parsed.Choices) == 0 {
		return "", 0, 0, fmt.Errorf("empty choices")
	}
	out := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if out == "" {
		return "", 0, 0, fmt.Errorf("empty content")
	}
	return out, parsed.Usage.PromptTokens, parsed.Usage.CompletionTokens, nil
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
