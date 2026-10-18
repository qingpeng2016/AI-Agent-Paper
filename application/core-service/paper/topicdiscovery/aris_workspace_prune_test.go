package topicdiscovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClearResearchLitInvokeWorkspace(t *testing.T) {
	root := t.TempDir()
	paperRoot := filepath.Join(root, "paper")
	if err := os.MkdirAll(filepath.Join(paperRoot, "literature"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paperRoot, "literature", "a.pdf"), []byte("pdf"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paperRoot, "LITERATURE_REVIEW.md"), []byte("# big"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(paperRoot, ".paper", "invoke", "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paperRoot, ".paper", "invoke", "logs", "keep.log"), []byte("log"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paperRoot, ".paper", "corpus.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paperRoot, "RESEARCH_BRIEF.md"), []byte("# brief"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := defaultARISIdeaDiscoveryConfig()
	cfg.ClearResearchLitInvokeWorkspace(paperRoot)

	if _, err := os.Stat(filepath.Join(paperRoot, "literature")); !os.IsNotExist(err) {
		t.Fatal("literature should be removed")
	}
	if _, err := os.Stat(filepath.Join(paperRoot, "LITERATURE_REVIEW.md")); !os.IsNotExist(err) {
		t.Fatal("LITERATURE_REVIEW.md should be removed")
	}
	if _, err := os.Stat(filepath.Join(paperRoot, ".paper", "corpus.json")); !os.IsNotExist(err) {
		t.Fatal("corpus.json should be removed")
	}
	if _, err := os.Stat(filepath.Join(paperRoot, ".paper", "invoke", "logs", "keep.log")); err != nil {
		t.Fatal("invoke logs should be kept")
	}
	if _, err := os.Stat(filepath.Join(paperRoot, "RESEARCH_BRIEF.md")); err != nil {
		t.Fatal("RESEARCH_BRIEF should be kept until rewritten")
	}
}
