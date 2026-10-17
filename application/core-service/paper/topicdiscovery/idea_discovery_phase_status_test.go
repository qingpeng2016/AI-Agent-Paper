package topicdiscovery

import "testing"

func TestPhaseStatusInMeta_legacyBoolAndStrings(t *testing.T) {
	meta := map[string]any{"research-lit": true, "idea-creator": false}
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	if got := phaseStatusInMeta(meta, "research-lit"); got != IdeaDiscoveryPhaseExecuting {
		t.Fatalf("research-lit: %q", got)
	}
	if got := phaseStatusInMeta(meta, "idea-creator"); got != IdeaDiscoveryPhaseNotStarted {
		t.Fatalf("idea-creator: %q", got)
	}
	meta["novelty-check"] = IdeaDiscoveryPhaseAwaitingConfirm
	if got := phaseStatusInMeta(meta, "novelty-check"); got != IdeaDiscoveryPhaseAwaitingConfirm {
		t.Fatalf("novelty-check: %q", got)
	}
}

func TestIdeaDiscoveryNeedsUserConfirm_respectsExecuting(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	meta := map[string]any{"idea-creator": IdeaDiscoveryPhaseExecuting}
	if ideaDiscoveryNeedsUserConfirm(meta, "idea-creator", cfg) {
		t.Fatal("executing should not need confirm")
	}
	meta["idea-creator"] = IdeaDiscoveryPhaseAwaitingConfirm
	if !ideaDiscoveryNeedsUserConfirm(meta, "idea-creator", cfg) {
		t.Fatal("awaiting_confirm should need confirm")
	}
}
