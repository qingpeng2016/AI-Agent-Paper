package topicdiscovery

import (
	"strings"
	"testing"
)

func TestBuildIdeaDiscoveryUIGroups_defaultYamlShape(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	groups := BuildIdeaDiscoveryUIGroups(cfg)
	if len(groups) != 5 {
		t.Fatalf("groups=%d want 5 (each phase pause:true)", len(groups))
	}
	if groups[0].LabelZh != "文献调研" || len(groups[0].PhaseIDs) != 1 {
		t.Fatalf("g0=%+v", groups[0])
	}
	if groups[1].LabelZh != "选题大纲" {
		t.Fatalf("g1=%+v", groups[1])
	}
}

func TestBuildIdeaDiscoveryUIGroups_threeFalseInRow(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	cfg.CheckpointPhases = normalizeIdeaDiscoveryCheckpointPhases(map[string]IdeaDiscoveryCheckpointPhase{
		"research-lit":             {Pause: false, LabelZh: "文献调研"},
		"idea-creator":             {Pause: false, LabelZh: "选题大纲"},
		"novelty-check":            {Pause: false, LabelZh: "新颖性检验"},
		"research-review":          {Pause: true, LabelZh: "外部审查"},
		"research-refine-pipeline": {Pause: true, LabelZh: "文献综述与实验规划"},
	})
	groups := BuildIdeaDiscoveryUIGroups(cfg)
	if len(groups) != 3 {
		t.Fatalf("groups=%d want 3", len(groups))
	}
	want := "文献调研、选题大纲、新颖性检验"
	if groups[0].LabelZh != want {
		t.Fatalf("g0=%q want %q", groups[0].LabelZh, want)
	}
}

func TestUIGroupIndexForPhase(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	groups := BuildIdeaDiscoveryUIGroups(cfg)
	if uiGroupIndexForPhase(groups, "research-lit") != 0 {
		t.Fatal("research-lit in group 0")
	}
	if uiGroupIndexForPhase(groups, "idea-creator") != 1 {
		t.Fatal("idea-creator in group 1")
	}
	if uiGroupIndexForPhase(groups, "novelty-check") != 2 {
		t.Fatal("novelty-check in group 2")
	}
}

func TestIdeaDiscoveryCheckpointDetail_usesPhaseLabelNotMergedGroup(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	cfg.CheckpointPhases = normalizeIdeaDiscoveryCheckpointPhases(map[string]IdeaDiscoveryCheckpointPhase{
		"research-lit":  {Pause: false, LabelZh: "文献调研"},
		"idea-creator":  {Pause: false, LabelZh: "选题大纲"},
		"novelty-check": {Pause: true, LabelZh: "新颖性检验"},
	})
	_ = cfg
	got := IdeaDiscoveryCheckpointDetail("research-lit")
	if strings.Contains(got, "文献调研、选题大纲") {
		t.Fatalf("checkpoint detail must not use merged UI label: %q", got)
	}
	if !strings.Contains(got, "文献调研") || !strings.Contains(got, "research-lit") {
		t.Fatalf("want single phase label: %q", got)
	}
}
