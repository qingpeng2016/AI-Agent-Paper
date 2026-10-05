package product

import (
	"testing"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

func TestExpandCardFeatureLine(t *testing.T) {
	p := &entity.Products{LimitTokens: 20_000_000, RPMLimit: 120}
	got := expandCardFeatureLine("{limit_tokens} tokens / 月", p)
	if got != "20M tokens / 月" {
		t.Fatalf("limit_tokens: got %q", got)
	}
	got = expandCardFeatureLine("{rpm_limit} RPM", p)
	if got != "120 RPM" {
		t.Fatalf("rpm_limit: got %q", got)
	}
}
