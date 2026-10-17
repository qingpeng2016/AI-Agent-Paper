package topicdiscovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadRunStateResumeTarget_donePhasesAdvanceToPending(t *testing.T) {
	dir := t.TempDir()
	runID := "test-run"
	path := mustIdeaDiscoveryConfig().RunStateFileAbs(dir, runID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	payload := `{
  "phases": [
    {"phase": "research-lit", "status": "done"},
    {"phase": "idea-creator", "status": "done"},
    {"phase": "novelty-check", "status": "pending"},
    {"phase": "research-review", "status": "pending"},
    {"phase": "research-refine-pipeline", "status": "pending"}
  ]
}`
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	phase, complete, err := readRunStateResumeTarget(dir, runID)
	if err != nil {
		t.Fatal(err)
	}
	if complete {
		t.Fatal("expected incomplete run")
	}
	if phase != "novelty-check" {
		t.Fatalf("resume phase = %q, want novelty-check", phase)
	}
}

func TestRunStatePhaseNeedsWork(t *testing.T) {
	for _, tc := range []struct {
		status string
		want   bool
	}{
		{"done", false},
		{"accepted", false},
		{"pending", true},
		{"running", true},
		{"failed", true},
	} {
		if got := runStatePhaseNeedsWork(tc.status); got != tc.want {
			t.Errorf("runStatePhaseNeedsWork(%q) = %v, want %v", tc.status, got, tc.want)
		}
	}
}
