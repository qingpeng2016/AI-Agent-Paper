package topicdiscovery

import "unicode/utf8"

const maxLLMCallLogJSONBytes = 512 * 1024

func truncateUTF8(s string, maxRunes int) string {
	if maxRunes <= 0 || s == "" {
		return s
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxRunes]) + "…(truncated)"
}

func truncateBytes(s string, maxBytes int) string {
	if maxBytes <= 0 || len(s) <= maxBytes {
		return s
	}
	if maxBytes <= 3 {
		return s[:maxBytes]
	}
	b := []byte(s)
	if len(b) <= maxBytes {
		return s
	}
	return string(b[:maxBytes-3]) + "..."
}

func clipJSONForMySQL(raw string) string {
	return truncateBytes(raw, maxLLMCallLogJSONBytes)
}
