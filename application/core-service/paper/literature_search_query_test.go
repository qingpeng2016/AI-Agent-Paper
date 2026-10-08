package paper

import (
	"strings"
	"testing"
)

func TestExtractLiteratureSearchQuery_fromKeywordsLine(t *testing.T) {
	direction := `研究内容：研究重度抑郁症（MDD）患者脑功能网络拓扑是否相对健康对照存在稳定、可重复的异常。
关键词：major depressive disorder, resting-state fMRI, functional connectivity, complex brain network, graph theory`
	q := ExtractLiteratureSearchQuery(direction)
	if !strings.Contains(q, "major depressive disorder") {
		t.Fatalf("missing MDD phrase: %q", q)
	}
	if !strings.Contains(q, "graph theory") {
		t.Fatalf("missing graph theory: %q", q)
	}
	if strings.Contains(q, "研究内容") {
		t.Fatalf("should not contain Chinese section header: %q", q)
	}
}

func TestExtractLiteratureSearchQuery_englishFallback(t *testing.T) {
	direction := "Topic about resting-state fMRI and graph theory in neuroscience."
	q := ExtractLiteratureSearchQuery(direction)
	if q == "" {
		t.Fatal("expected non-empty query")
	}
	if !strings.Contains(strings.ToLower(q), "fmri") {
		t.Fatalf("expected fMRI terms: %q", q)
	}
}
