package topicdiscovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSeedIdeaDiscoveryRunStateFromWorkspace(t *testing.T) {
	dir := t.TempDir()
	runID := "test-run-20261018"
	cfg := defaultARISIdeaDiscoveryConfig()
	report := filepath.Join(dir, cfg.Artifacts.IdeaReport)
	if err := os.MkdirAll(filepath.Dir(report), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "# Report\n## 📚 文献景观\n\n" + strings.Repeat("x", 512) + "\n## 💡 想法候选池\n\ncontent"
	if err := os.WriteFile(report, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := seedIdeaDiscoveryRunStateFromWorkspace(dir, runID); err != nil {
		t.Fatal(err)
	}
	if !ideaDiscoveryRunStateFileExists(dir, runID) {
		t.Fatal("run_state should be seeded")
	}
	phase, _, err := readRunStateResumeTarget(dir, runID)
	if err != nil {
		t.Fatal(err)
	}
	if phase != "novelty-check" {
		t.Fatalf("pending phase %q want novelty-check", phase)
	}
}
