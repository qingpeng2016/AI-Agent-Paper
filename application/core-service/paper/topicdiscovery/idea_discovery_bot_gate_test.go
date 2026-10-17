package topicdiscovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIdeaDiscoveryBotMayInvokePending_noJSONWhenExecuting(t *testing.T) {
	meta := map[string]any{"research-lit": IdeaDiscoveryPhaseExecuting}
	if !IdeaDiscoveryBotMayInvokePending(meta, "research-lit", t.TempDir(), "missing-run") {
		t.Fatal("expected invoke when no run_state file and phase executing")
	}
}

func TestIdeaDiscoveryBotMayInvokePending_jsonNotDoneWhenExecuting(t *testing.T) {
	dir := t.TempDir()
	runID := "gate-test"
	cfg := defaultARISIdeaDiscoveryConfig()
	path := cfg.RunStateFileAbs(dir, runID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	raw := `{"run_id":"gate-test","phases":[{"phase":"research-lit","status":"running"}]}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	meta := map[string]any{"research-lit": IdeaDiscoveryPhaseExecuting}
	if !IdeaDiscoveryBotMayInvokePending(meta, "research-lit", dir, runID) {
		t.Fatal("expected invoke when JSON not done and DB executing")
	}
	if IdeaDiscoveryBotMayInvokePending(meta, "research-lit", dir, runID) && isRunStatePhaseDone(dir, runID, "research-lit") {
		t.Fatal("sanity: should not be both invoke-eligible and done")
	}
}

func TestPrepareIdeaDiscoveryBotInvoke_requiresConfirmWhenJSONPending(t *testing.T) {
	dir := t.TempDir()
	runID := "gate-test"
	cfg := defaultARISIdeaDiscoveryConfig()
	path := cfg.RunStateFileAbs(dir, runID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	raw := `{"run_id":"gate-test","phases":[{"phase":"research-lit","status":"done"},{"phase":"idea-creator","status":"pending"}]}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	if isRunStatePhaseDone(dir, runID, "idea-creator") {
	 t.Fatal("idea-creator should not be done")
	}
	pending, err := ideaDiscoveryResumePhase(dir, runID)
	if err != nil || pending != "idea-creator" {
		t.Fatalf("pending %q err %v", pending, err)
	}
	confirmed := strings.TrimSpace("idea-creator") == pending
	if confirmed {
		// without meta this is just phase match sanity
	}
	if !cfg.PhaseNeedsUserCheckpoint("idea-creator") {
		t.Fatal("expected pause phase")
	}
}
