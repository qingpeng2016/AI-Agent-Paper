package topicdiscovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIdeaDiscoveryPhasedPhaseSucceeded_researchLitLiteratureReview(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "LITERATURE_REVIEW.md")
	if err := os.WriteFile(path, []byte("# lit\n"+string(make([]byte, 300))), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := defaultARISIdeaDiscoveryConfig()
	if !ideaDiscoveryPhasedPhaseSucceeded(root, "any-run", "research-lit", cfg) {
		t.Fatal("expected LITERATURE_REVIEW.md to mark research-lit done")
	}
}

func TestIdeaDiscoveryPhasedPhaseSucceeded_researchLitReport(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "RESEARCH_LIT_REPORT.md")
	if err := os.WriteFile(path, []byte("# report\n"+string(make([]byte, 300))), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := defaultARISIdeaDiscoveryConfig()
	if !ideaDiscoveryPhasedPhaseSucceeded(root, "any-run", "research-lit", cfg) {
		t.Fatal("expected RESEARCH_LIT_REPORT.md to mark research-lit done")
	}
}

func TestParseIdeaDiscoveryPhasedInvokeResult(t *testing.T) {
	ok, _, err := parseIdeaDiscoveryPhasedInvokeResult(json.RawMessage(`{"result":"success"}`))
	if err != nil || !ok {
		t.Fatalf("success: ok=%v err=%v", ok, err)
	}
	ok, reason, err := parseIdeaDiscoveryPhasedInvokeResult(json.RawMessage(`{"result":"false","reason":"token limit"}`))
	if err != nil || ok || reason != "token limit" {
		t.Fatalf("false: ok=%v reason=%q err=%v", ok, reason, err)
	}
	_, _, err = parseIdeaDiscoveryPhasedInvokeResult(json.RawMessage(`{"ideas":[]}`))
	if err == nil {
		t.Fatal("expected error for missing result")
	}
}

func TestBuildIdeaDiscoveryPhaseCLIPrompt_includesResultHint(t *testing.T) {
	p := buildIdeaDiscoveryPhaseCLIPrompt("research-lit", "FOO: bar", defaultARISIdeaDiscoveryConfig())
	if !strings.Contains(p, `"result":"success"`) || !strings.Contains(p, `"result":"false"`) {
		t.Fatalf("prompt missing result JSON hint: %q", p)
	}
}

func TestIdeaDiscoveryPhasedBotMayRetryPhase_never(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	cfg.Platform.AutoProceed = true
	if IdeaDiscoveryPhasedBotMayRetryPhase(cfg) {
		t.Fatal("phased bot must not re-queue same phase on missing artifacts")
	}
}
