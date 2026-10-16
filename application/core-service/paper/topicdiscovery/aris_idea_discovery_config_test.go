package topicdiscovery

import (
	"strings"
	"testing"
)

func TestDefaultARISIdeaDiscoveryConfigMatchesSkill(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	if cfg.Artifacts.IdeaReport != "idea-stage/IDEA_REPORT.md" {
		t.Fatalf("idea_report: %s", cfg.Artifacts.IdeaReport)
	}
	if cfg.Gate.Report != cfg.Artifacts.IdeaReport {
		t.Fatalf("gate report should match idea report")
	}
	if cfg.Skill.AutoProceed {
		t.Fatal("platform default AUTO_PROCEED is false (human checkpoint between phases)")
	}
	if !cfg.Platform.InterStepHumanCheckpoint {
		t.Fatal("inter_step_human_checkpoint should be true")
	}
	if cfg.Skill.Sources != "local" {
		t.Fatalf("sources: %s", cfg.Skill.Sources)
	}
	if cfg.RunState.ProvisionalAdvances {
		t.Fatal("native default provisional_advances must be false")
	}
	line := cfg.formatSkillArgumentsLine("/tmp/paper", "dir", "NeurIPS")
	if !strings.Contains(line, "AUTO_PROCEED: false") || !strings.Contains(line, "sources: local") {
		t.Fatalf("args line: %s", line)
	}
	if !strings.Contains(line, "paper library:") || !strings.Contains(line, "literature") {
		t.Fatalf("args should include paper library path: %s", line)
	}
	if !strings.Contains(line, "composed: idea-stage/IDEA_REPORT.md") {
		t.Fatalf("args line: %s", line)
	}
}
