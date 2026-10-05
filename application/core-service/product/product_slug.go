package product

import "strings"

// slug → sku_code（与前端 productRoutes 保持一致）
var slugToSKU = map[string]string{
	"gpt-go":          "OAI-GO-M",
	"chatgpt-plus":    "OAI-PLUS-M",
	"chatgpt-pro-5x":  "OAI-PRO5-M",
	"chatgpt-pro-20x": "OAI-PRO20-M",
	"claude-pro":      "ANT-PRO-M",
	"claude-max-5x":   "ANT-MAX-M",
	"grok-super":      "XAI-GROK-M",
	"gemini-pro":      "GEM-PRO-M",
	"cursor-pro":      "CUR-PRO-M",
	"cursor-pro-plus": "CUR-PRO-M",
	"perplexity-pro":  "PPX-PRO-M",
}

var skuToSlug = map[string]string{
	"OAI-GO-M":    "gpt-go",
	"OAI-PLUS-M":  "chatgpt-plus",
	"OAI-PRO5-M":  "chatgpt-pro-5x",
	"OAI-PRO20-M": "chatgpt-pro-20x",
	"ANT-PRO-M":   "claude-pro",
	"ANT-MAX-M":   "claude-max-5x",
	"XAI-GROK-M":  "grok-super",
	"GEM-PRO-M":   "gemini-pro",
	"CUR-PRO-M":   "cursor-pro",
	"PPX-PRO-M":   "perplexity-pro",
}

func skuFromSlug(slug string) string {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if sku, ok := slugToSKU[slug]; ok {
		return sku
	}
	return ""
}

func canonicalSlugForSKU(sku string) string {
	if s, ok := skuToSlug[sku]; ok {
		return s
	}
	return strings.ToLower(strings.ReplaceAll(sku, "_", "-"))
}
