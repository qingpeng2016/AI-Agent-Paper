package apikey

import "strings"

// MaskedFromHash 列表脱敏（仅存 SHA-256 哈希时使用）。
func MaskedFromHash(hash string) string {
	hash = strings.TrimSpace(hash)
	if len(hash) < 12 {
		return "ap-••••••••••••"
	}
	return hash[:4] + "••••••" + hash[len(hash)-4:]
}
