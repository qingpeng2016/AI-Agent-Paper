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
	if cfg.Platform.ArxivMaxDownload != 5 {
		t.Fatalf("arxiv_max_download: %d", cfg.Platform.ArxivMaxDownload)
	}
	if cfg.Platform.MaxArisPerBot != 1 {
		t.Fatalf("max_aris_per_bot: %d", cfg.Platform.MaxArisPerBot)
	}
	if !cfg.Platform.Resume {
		t.Fatal("resume should default true")
	}
	in := topicRunInput{
		Direction:   "MDD rs-fMRI graph",
		Description: "研究内容：MDD 脑网络",
		Keywords:    []string{"major depressive disorder", "graph theory"},
	}
	line := cfg.FormatIdeaDiscoveryArgsLine(in, "NeurIPS", "en", "run-abc", true)
	if !strings.Contains(line, "AUTO_PROCEED: false") {
		t.Fatalf("args line: %s", line)
	}
	if !strings.Contains(line, "— venue: NeurIPS —") || !strings.Contains(line, "— sources: web —") {
		t.Fatalf("args line: %s", line)
	}
	if !strings.Contains(line, "ARXIV_DOWNLOAD: true") || !strings.Contains(line, "ARXIV_MAX_DOWNLOAD: 5") {
		t.Fatalf("args line: %s", line)
	}
	if !strings.Contains(line, "— language: en") || !strings.Contains(line, "— resume run-abc") {
		t.Fatalf("args line: %s", line)
	}
	if !strings.HasPrefix(line, `"`) {
		t.Fatalf("quoted direction: %s", line)
	}
	if cfg.InvokeLogsDirAbs("/tmp/paper") != filepath.Join("/tmp/paper", ".paper", "invoke", "logs") {
		t.Fatalf("logs abs: %s", cfg.InvokeLogsDirAbs("/tmp/paper"))
	}
}

func TestFormatIdeaDiscoveryArgsLine_autoProceedTrue(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	cfg.Platform.AutoProceed = true
	in := topicRunInput{Direction: "dir"}
	line := cfg.FormatIdeaDiscoveryArgsLine(in, "NeurIPS", "zh", "", false)
	if !strings.Contains(line, "AUTO_PROCEED: true") {
		t.Fatalf("args line: %s", line)
	}
}
