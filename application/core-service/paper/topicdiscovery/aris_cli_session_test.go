package topicdiscovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClearGlobalClaudeSessionsReferencingWorkspace(t *testing.T) {
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

	clearGlobalClaudeSessionsReferencingWorkspace(paperRoot)

	if _, err := os.Stat(matchPath); !os.IsNotExist(err) {
		t.Fatal("matching global session should be removed")
	}
	if _, err := os.Stat(otherPath); err != nil {
		t.Fatal("unrelated global session should remain")
	}
}

func TestClearARISCLIWorkspaceSession_localDir(t *testing.T) {
	root := t.TempDir()
	paperRoot := filepath.Join(root, "paper")
	sess := filepath.Join(paperRoot, ".claude", "sessions", "session-x.json")
	if err := os.MkdirAll(filepath.Dir(sess), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sess, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	clearARISCLIWorkspaceSession(paperRoot)
	if _, err := os.Stat(filepath.Join(paperRoot, ".claude")); !os.IsNotExist(err) {
		t.Fatal(".claude under paperRoot should be removed")
	}
	if strings.Contains(root, " ") {
		t.Skip("unused")
	}
}
