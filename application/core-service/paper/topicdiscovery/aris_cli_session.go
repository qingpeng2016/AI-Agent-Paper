package topicdiscovery

import (
	"os"
	"path/filepath"
	"strings"
)

// clearARISCLIWorkspaceSession 删除 paperRoot 下 ARIS/Claude 持久化会话（与 platform.resume 无关）。
// 另清理 bot 用户 $HOME/.claude/sessions 里仍引用该 workspace 路径的旧 session（删 inout 目录无法去掉这部分）。
func clearARISCLIWorkspaceSession(paperRoot string) {
	paperRoot = filepath.Clean(strings.TrimSpace(paperRoot))
	if paperRoot == "" {
		return
	}
	for _, rel := range []string{".claude/sessions", ".claude"} {
		_ = os.RemoveAll(filepath.Join(paperRoot, filepath.FromSlash(rel)))
	}
	clearGlobalClaudeSessionsReferencingWorkspace(paperRoot)
}

func clearGlobalClaudeSessionsReferencingWorkspace(paperRoot string) {
	absRoot, err := filepath.Abs(paperRoot)
	if err != nil {
		absRoot = paperRoot
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return
	}
	sessionsDir := filepath.Join(home, ".claude", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "session-") || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(sessionsDir, e.Name())
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		text := string(b)
		if strings.Contains(text, absRoot) || strings.Contains(text, paperRoot) {
			_ = os.Remove(path)
		}
	}
}
