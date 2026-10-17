package topicdiscovery

import "testing"

func TestBuildIdeaDiscoveryUIGroups_defaultYamlShape(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	groups := BuildIdeaDiscoveryUIGroups(cfg)
	if len(groups) != 3 {
		t.Fatalf("groups=%d want 3", len(groups))
	}
	if groups[0].LabelZh != "文献调研、选题大纲" {
		t.Fatalf("g0 label=%q", groups[0].LabelZh)
	}
	if len(groups[0].PhaseIDs) != 2 {
		t.Fatalf("g0 phases=%v", groups[0].PhaseIDs)
	}
}

func TestBuildIdeaDiscoveryUIGroups_threeFalseInRow(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	cfg.CheckpointPhases = normalizeIdeaDiscoveryCheckpointPhases(map[string]IdeaDiscoveryCheckpointPhase{
		"research-lit":  {Pause: false, LabelZh: "文献调研"},
		"idea-creator":  {Pause: false, LabelZh: "选题大纲"},
		"novelty-check": {Pause: false, LabelZh: "新颖性检验"},
		"research-review":          {Pause: true, LabelZh: "外部审查"},
		"research-refine-pipeline": {Pause: true, LabelZh: "文献综述与实验规划"},
	})
	groups := BuildIdeaDiscoveryUIGroups(cfg)
	if len(groups) != 2 {
		t.Fatalf("groups=%d want 2", len(groups))
	}
	want := "文献调研、选题大纲、新颖性检验"
	if groups[0].LabelZh != want {
		t.Fatalf("g0=%q want %q", groups[0].LabelZh, want)
	}
}

func TestUIGroupIndexForPhase(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	groups := BuildIdeaDiscoveryUIGroups(cfg)
	if uiGroupIndexForPhase(groups, "idea-creator") != 0 {
		t.Fatal("idea-creator in group 0")
	}
	if uiGroupIndexForPhase(groups, "novelty-check") != 1 {
		t.Fatal("novelty-check in group 1")
	}
}
