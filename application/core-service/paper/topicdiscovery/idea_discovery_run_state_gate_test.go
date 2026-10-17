package topicdiscovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIdeaDiscoveryRunStateFileExists(t *testing.T) {
	dir := t.TempDir()
	runID := "mdd_brain_network_2026-10-17"
	if ideaDiscoveryRunStateFileExists(dir, runID) {
		t.Fatal("missing file should be false")
	}
	runs := filepath.Join(dir, ".aris", "runs")
	if err := os.MkdirAll(runs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runs, runID+".json"), []byte(`{"phases":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if !ideaDiscoveryRunStateFileExists(dir, runID) {
		t.Fatal("file present should be true")
	}
}

func TestPhaseNeedsUserCheckpoint_noveltyOnly(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	if cfg.PhaseNeedsUserCheckpoint("research-lit") {
		t.Fatal("research-lit pause:false")
	}
	if !cfg.PhaseNeedsUserCheckpoint("novelty-check") {
		t.Fatal("novelty-check pause:true")
	}
}
