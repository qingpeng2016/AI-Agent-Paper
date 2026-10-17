package topicdiscovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefreshRunStateMarksIdeaCreatorDoneFromReport(t *testing.T) {
	dir := t.TempDir()
	runID := "test-run-20261018"
	cfg := defaultARISIdeaDiscoveryConfig()
	report := filepath.Join(dir, cfg.Artifacts.IdeaReport)
	if err := os.MkdirAll(filepath.Dir(report), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "# Report\n## 📚 文献景观\n\n" + strings.Repeat("a", 520) + "\n## 💡 初步想法\nIdea 1.1 test\n"
	if err := os.WriteFile(report, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeIdeaDiscoveryRunStateFromReport(dir, runID); err != nil {
		t.Fatal(err)
	}
	phase, _, err := readRunStateResumeTarget(dir, runID)
	if err != nil {
		t.Fatal(err)
	}
	if phase != "novelty-check" {
		t.Fatalf("initial pending %q want novelty-check", phase)
	}
	// Simulate stale seed: only research-lit done.
	path := cfg.RunStateFileAbs(dir, runID)
	stale := `{"run_id":"test","seeded_by":"ai-agent-paper","phases":[{"phase":"research-lit","status":"done"},{"phase":"idea-creator","status":"pending"}]}`
	if err := os.WriteFile(path, []byte(stale), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := refreshIdeaDiscoveryRunStateFromWorkspace(dir, runID); err != nil {
		t.Fatal(err)
	}
	phase, _, err = readRunStateResumeTarget(dir, runID)
	if err != nil {
		t.Fatal(err)
	}
	if phase != "novelty-check" {
		t.Fatalf("after refresh pending %q want novelty-check", phase)
	}
}
