package step

import (
	"encoding/json"
	"testing"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

func TestIdeaDiscoveryRuntimeState_stageAndOtherRoundTrip(t *testing.T) {
	step := &entity.PaperOutputTopicStep{StageCode: "idea_discovery"}
	meta := map[string]any{
		"async_llm":                  true,
		"native_run_id":              "run-abc",
		"research-lit":               "executing",
		"idea-creator":               "not_started",
		"novelty-check":              "not_started",
		"research-review":            "not_started",
		"research-refine-pipeline":   "not_started",
		"pipeline":                   "aris_idea_discovery_native",
	}
	PersistStepExtra(step, meta)

	var rt map[string]any
	if err := json.Unmarshal(step.RuntimeState, &rt); err != nil {
		t.Fatal(err)
	}
	if _, ok := rt["research-lit"]; ok {
		t.Fatal("flat phase key should not be stored at top level")
	}
	if rt["async_llm"] != nil {
		t.Fatal("async_llm should be under other")
	}
	other, _ := rt["other"].(map[string]any)
	if other == nil || other["async_llm"] != true {
		t.Fatalf("other.async_llm: %#v", rt["other"])
	}
	stage, _ := rt["stage"].(map[string]any)
	if stage == nil {
		t.Fatal("missing stage")
	}
	e1, _ := stage["1"].(map[string]any)
	if e1["name"] != "research-lit" || e1["status"] != "executing" {
		t.Fatalf("stage[1]: %#v", e1)
	}
	e2, _ := stage["2"].(map[string]any)
	if e2["name"] != "idea-creator" {
		t.Fatalf("stage[2]: %#v", e2)
	}

	loaded := LoadStepExtra(step)
	if loaded["async_llm"] != true {
		t.Fatal("load should flatten async_llm")
	}
	if loaded["research-lit"] != "executing" {
		t.Fatal("load should flatten phase status")
	}
}
