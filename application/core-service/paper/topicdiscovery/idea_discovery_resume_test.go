package topicdiscovery

import (
	"os"
	"path/filepath"
	"testing"
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
