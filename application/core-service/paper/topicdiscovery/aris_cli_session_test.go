package topicdiscovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClearGlobalSessionsClaude(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	paperRoot := filepath.Join(home, "inout", "user_1", "manuscript_1", "run_1", "paper")
	sessionsDir := filepath.Join(home, ".claude", "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	matchPath := filepath.Join(sessionsDir, "session-test.json")
	otherPath := filepath.Join(sessionsDir, "session-other.json")
	if err := os.WriteFile(matchPath, []byte(`{"cwd":"`+paperRoot+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(otherPath, []byte(`{"cwd":"/tmp/elsewhere"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	clearGlobalSessionsClaude(paperRoot, sessionsDir)

	if _, err := os.Stat(matchPath); !os.IsNotExist(err) {
		t.Fatal("matching global session should be removed")
	}
	if _, err := os.Stat(otherPath); err != nil {
		t.Fatal("unrelated global session should remain")
	}
}

func TestClearGlobalSessionsCodexRecursive(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	paperRoot := filepath.Join(home, "inout", "user_1", "manuscript_1", "run_1", "paper")
	nested := filepath.Join(home, ".codex", "sessions", "2026", "01")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	matchPath := filepath.Join(nested, "thread-abc.json")
	if err := os.WriteFile(matchPath, []byte(`{"workspace":"`+paperRoot+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	clearGlobalSessionsCodex(paperRoot, filepath.Join(home, ".codex", "sessions"))

	if _, err := os.Stat(matchPath); !os.IsNotExist(err) {
		t.Fatal("codex nested session should be removed")
	}
}

func TestClearCLIWorkspaceSessions_localWorkspaceDirs(t *testing.T) {
	root := t.TempDir()
	paperRoot := filepath.Join(root, "paper")
	for _, rel := range []string{".claude/sessions/x.json", ".codex/state.json"} {
		p := filepath.Join(paperRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := defaultARISIdeaDiscoveryConfig()
	cfg.ClearCLIWorkspaceSessions(paperRoot)
	if _, err := os.Stat(filepath.Join(paperRoot, ".claude")); !os.IsNotExist(err) {
		t.Fatal(".claude should be removed")
	}
	if _, err := os.Stat(filepath.Join(paperRoot, ".codex")); !os.IsNotExist(err) {
		t.Fatal(".codex should be removed")
	}
}

func TestExpandConfigPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	got := expandConfigPath("~/foo/bar")
	want := filepath.Join(home, "foo", "bar")
	if got != want {
		t.Fatalf("expand: got %q want %q", got, want)
	}
}
