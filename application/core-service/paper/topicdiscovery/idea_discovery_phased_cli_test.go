package topicdiscovery

import (
	"os"
	"path/filepath"
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

func TestIdeaDiscoveryPhasedBotMayRetryPhase_never(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	cfg.Platform.AutoProceed = true
	if IdeaDiscoveryPhasedBotMayRetryPhase(cfg) {
		t.Fatal("phased bot must not re-queue same phase on missing artifacts")
	}
}
