package topicdiscovery

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIdeaDiscoveryShouldAppendResume_freshRun(t *testing.T) {
	dir := t.TempDir()
	if ideaDiscoveryShouldAppendResume(dir, "mdd-rsfmri-graph-topology-20261015") {
		t.Fatal("no run_state file should not resume")
	}
}

func TestIdeaDiscoveryShouldAppendResume_ignoresDBMeta(t *testing.T) {
	dir := t.TempDir()
	// DB 有进度但磁盘无 run_state → 仍不加 resume
	if ideaDiscoveryShouldAppendResume(dir, "run-1") {
		t.Fatal("resume must not depend on DB")
	}
}

func TestNewestIdeaDiscoveryRunStateID(t *testing.T) {
	dir := t.TempDir()
	cfg := defaultARISIdeaDiscoveryConfig()
	stateDir := filepath.Join(dir, ".aris", "runs")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldPath := cfg.RunStateFileAbs(dir, "older-run")
	newPath := cfg.RunStateFileAbs(dir, "mdd-rs-fmri-idea-discovery-20261017")
	if err := os.WriteFile(oldPath, []byte(`{"phases":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte(`{"phases":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Chtimes(oldPath, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour))
	got := newestIdeaDiscoveryRunStateID(dir)
	if got != "mdd-rs-fmri-idea-discovery-20261017" {
		t.Fatalf("newest run_id = %q", got)
	}
}

func TestIdeaDiscoveryPhaseNeedsUserCheckpoint(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	for _, tc := range []struct {
		phase string
		want  bool
	}{
		{"research-lit", false},
		{"idea-creator", false},
		{"novelty-check", true},
		{"research-review", true},
		{"research-refine-pipeline", true},
		{"", false},
	} {
		if got := cfg.PhaseNeedsUserCheckpoint(tc.phase); got != tc.want {
			t.Fatalf("phase %q: got %v want %v", tc.phase, got, tc.want)
		}
	}
	cfg.CheckpointPhases = normalizeIdeaDiscoveryCheckpointPhases(map[string]IdeaDiscoveryCheckpointPhase{
		"research-lit": {Pause: true, LabelZh: "文献调研"},
	})
	if !cfg.PhaseNeedsUserCheckpoint("research-lit") {
		t.Fatal("checkpoint_phases research-lit pause:true should pause")
	}
	if cfg.PhaseLabelZh("research-lit") != "文献调研" {
		t.Fatalf("label_zh: %q", cfg.PhaseLabelZh("research-lit"))
	}
	if cfg.PhaseNeedsUserCheckpoint("idea-creator") {
		t.Fatal("idea-creator default false should not pause")
	}
}

func TestIdeaDiscoveryShouldAppendResume_runStateFile(t *testing.T) {
	dir := t.TempDir()
	cfg := defaultARISIdeaDiscoveryConfig()
	runID := "mdd-rsfmri-graph-topology-20261015"
	stateDir := filepath.Join(dir, ".aris", "runs")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	state := `{"phases":[{"phase":"research-lit","status":"pending"}]}`
	path := cfg.RunStateFileAbs(dir, runID)
	if err := os.WriteFile(path, []byte(state), 0o644); err != nil {
		t.Fatal(err)
	}
	if !ideaDiscoveryShouldAppendResume(dir, runID) {
		t.Fatal("run_state file present should resume")
	}
}
