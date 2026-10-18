package topicdiscovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNukeClaudeProjectsForPaperRoot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	paperRoot := "/www/wwwroot/AI-Agent-Paper/inout/user_1/manuscript_1/run_1/paper"
	projectsRoot := filepath.Join(home, ".claude", "projects")
	dirName := "-www-wwwroot-AI-Agent-Paper-inout-user-1-manuscript-1-run-1-paper"
	projectDir := filepath.Join(projectsRoot, dirName)
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "thread.jsonl"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if n := nukeClaudeProjectsForPaperRoot(paperRoot); n != 1 {
		t.Fatalf("expected 1 project dir removed, got %d", n)
	}
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		t.Fatal("project dir should be removed")
	}
}

func TestClaudeProjectDirMatchSlug(t *testing.T) {
	got := claudeProjectDirMatchSlug("/www/wwwroot/AI-Agent-Paper/inout/user_1/manuscript_1/run_2/paper")
	want := "user-1-manuscript-1-run-2-paper"
	if got != want {
		t.Fatalf("slug %q want %q", got, want)
	}
}
