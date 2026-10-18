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

func TestIdeaDiscoveryPhasedPhaseInvokeSucceeded_researchLitFile(t *testing.T) {
	root := t.TempDir()
	ok, _, err := ideaDiscoveryPhasedPhaseInvokeSucceeded(root, "research-lit", json.RawMessage(`{"result":"success"}`))
	if err == nil || ok {
		t.Fatalf("missing file: ok=%v err=%v", ok, err)
	}
	path := filepath.Join(root, researchLitLandscapeRel)
	if err := os.WriteFile(path, []byte("# landscape\n"+string(make([]byte, 80))), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, _, err = ideaDiscoveryPhasedPhaseInvokeSucceeded(root, "research-lit", json.RawMessage(`{"ideas":[]}`))
	if err != nil || !ok {
		t.Fatalf("with file: ok=%v err=%v", ok, err)
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

func TestBuildIdeaDiscoveryPhaseCLIPrompt_skillOnly(t *testing.T) {
	p := buildIdeaDiscoveryPhaseCLIPrompt("research-lit", "FOO: bar", defaultARISIdeaDiscoveryConfig())
	if p != "/research-lit FOO: bar" {
		t.Fatalf("prompt: %q", p)
	}
	if strings.Contains(p, "AI-Agent-Paper phased invoke") {
		t.Fatalf("prompt must not append platform JSON hint: %q", p)
	}
}

func TestIdeaDiscoveryPhasedBotMayRetryPhase_never(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	cfg.Platform.AutoProceed = true
	if IdeaDiscoveryPhasedBotMayRetryPhase(cfg) {
		t.Fatal("phased bot must not re-queue same phase on missing artifacts")
	}
}
