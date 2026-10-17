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
	if cfg.Platform.Sources != "web" {
		t.Fatalf("sources: %s", cfg.Platform.Sources)
	}
	if !cfg.Platform.ArxivDownload {
		t.Fatal("arxiv_download should be true")
	}
	if cfg.Platform.MaxArisPerBot != 1 {
		t.Fatalf("max_aris_per_bot: %d", cfg.Platform.MaxArisPerBot)
	}
	line := cfg.formatSkillArgumentsLine("/tmp/paper", "dir", "NeurIPS", "en")
	if !strings.Contains(line, "AUTO_PROCEED: false") {
		t.Fatalf("args line: %s", line)
	}
	if !strings.Contains(line, "sources: web") || !strings.Contains(line, "ARXIV_DOWNLOAD: true") || !strings.Contains(line, "language: en") {
		t.Fatalf("args line: %s", line)
	}
	if cfg.InvokeLogsDirAbs("/tmp/paper") != filepath.Join("/tmp/paper", ".paper", "invoke", "logs") {
		t.Fatalf("logs abs: %s", cfg.InvokeLogsDirAbs("/tmp/paper"))
	}
}

func TestFormatSkillArgumentsLine_autoProceedTrue(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	cfg.Platform.AutoProceed = true
	line := cfg.formatSkillArgumentsLine("/tmp/paper", "dir", "NeurIPS", "en")
	if !strings.Contains(line, "AUTO_PROCEED: true") {
		t.Fatalf("args line: %s", line)
	}
}
