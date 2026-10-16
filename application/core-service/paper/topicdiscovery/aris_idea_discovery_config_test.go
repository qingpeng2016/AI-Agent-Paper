package topicdiscovery

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultARISIdeaDiscoveryConfig(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	if cfg.Platform.AutoProceed {
		t.Fatal("default auto_proceed is false")
	}
	if !cfg.Platform.PauseAfterRetrieve {
		t.Fatal("default pause_after_retrieve should be true")
	}
	if cfg.Platform.Sources != "local" {
		t.Fatalf("sources: %s", cfg.Platform.Sources)
	}
	if cfg.Platform.MaxArisPerBot != 1 {
		t.Fatalf("max_aris_per_bot: %d", cfg.Platform.MaxArisPerBot)
	}
	if cfg.Artifacts.IdeaReport != "idea-stage/IDEA_REPORT.md" {
		t.Fatalf("idea_report: %s", cfg.Artifacts.IdeaReport)
	}
	if cfg.Workspace.InvokeLogsDir != ".paper/invoke/logs" {
		t.Fatalf("invoke_logs_dir: %s", cfg.Workspace.InvokeLogsDir)
	}
	if cfg.InvokeLogsDirAbs("/tmp/paper") != filepath.Join("/tmp/paper", ".paper", "invoke", "logs") {
		t.Fatalf("logs abs: %s", cfg.InvokeLogsDirAbs("/tmp/paper"))
	}
	line := cfg.formatSkillArgumentsLine("/tmp/paper", "dir", "NeurIPS")
	if !strings.Contains(line, "AUTO_PROCEED: false") || !strings.Contains(line, "sources: local") {
		t.Fatalf("args line: %s", line)
	}
	if !strings.Contains(line, "paper library:") || !strings.Contains(line, "literature") {
		t.Fatalf("args should include paper library path: %s", line)
	}
}
