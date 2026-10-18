package topicdiscovery

import "testing"

func TestExecutingPhaseFromRuntimeState_queueRow(t *testing.T) {
	raw := []byte(`{
		"other": {"async_llm": true},
		"stage": {
			"1": {"name": "research-lit", "status": "executing"},
			"2": {"name": "idea-creator", "status": "not_started"},
			"3": {"name": "novelty-check", "status": "not_started"},
			"4": {"name": "research-review", "status": "not_started"},
			"5": {"name": "research-refine-pipeline", "status": "not_started"}
		}
	}`)
	if !runtimeStateOtherAsyncLLM(raw) {
		t.Fatal("expected async_llm")
	}
	if got := ExecutingPhaseFromRuntimeState(raw); got != "research-lit" {
		t.Fatalf("executing phase: got %q", got)
	}
}
