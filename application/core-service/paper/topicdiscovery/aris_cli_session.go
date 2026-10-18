package topicdiscovery

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// IdeaDiscoveryCLISessionCleanup 仅配置 bot 用户下全局 session 根目录（支持 ~/、$HOME/）；其余规则写死在代码中。
type IdeaDiscoveryCLISessionCleanup struct {
	Claude string `yaml:"claude"`
	Codex  string `yaml:"codex"`
}

func defaultCLISessionCleanup() IdeaDiscoveryCLISessionCleanup {
	return IdeaDiscoveryCLISessionCleanup{
		Claude: "~/.claude/sessions",
		Codex:  "~/.codex/sessions",
	}
}

func normalizeCLISessionCleanup(c *IdeaDiscoveryCLISessionCleanup) {
	if c == nil {
		return
	}
	def := defaultCLISessionCleanup()
	if strings.TrimSpace(c.Claude) == "" {
		c.Claude = def.Claude
	}
	if strings.TrimSpace(c.Codex) == "" {
		c.Codex = def.Codex
	}
}

func expandConfigPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Join(home, p[2:])
		}
	}
	if strings.HasPrefix(p, "$HOME/") || strings.HasPrefix(p, "$HOME\\") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Join(home, p[6:])
		}
	}
	return os.ExpandEnv(p)
}

// ClearCLIWorkspaceSessions 每次 invoke 前清理该 workspace 下的 CLI 会话目录，避免同 run 多次
// aris prompt 复用 Claude/Codex 会话导致上下文累积。
func (c ARISIdeaDiscoveryConfig) ClearCLIWorkspaceSessions(paperRoot string) {
	cleanup := c.CLISessionCleanup
	normalizeCLISessionCleanup(&cleanup)
	clearCLIWorkspaceSessions(paperRoot, cleanup)
}

func clearCLIWorkspaceSessions(paperRoot string, cleanup IdeaDiscoveryCLISessionCleanup) {
	paperRoot = filepath.Clean(strings.TrimSpace(paperRoot))
	if paperRoot == "" {
		return
	}
	// Project-local Claude session logs (e.g. …/paper/.claude/sessions) must not survive between invokes.
	_ = os.RemoveAll(filepath.Join(paperRoot, ".claude", "sessions"))
	for _, rel := range []string{".claude", ".codex"} {
		_ = os.RemoveAll(filepath.Join(paperRoot, rel))
	}
	clearGlobalSessionsClaude(paperRoot, expandConfigPath(cleanup.Claude))
	clearGlobalSessionsCodex(paperRoot, expandConfigPath(cleanup.Codex))
}

func clearGlobalSessionsClaude(paperRoot, sessionsDir string) {
	if sessionsDir == "" {
		return
	}
	matches, err := filepath.Glob(filepath.Join(sessionsDir, "session-*.json"))
	if err != nil {
		return
	}
	removeSessionFilesReferencingWorkspace(paperRoot, matches)
}

func clearGlobalSessionsCodex(paperRoot, sessionsDir string) {
	if sessionsDir == "" {
		return
	}
	var paths []string
	_ = filepath.WalkDir(sessionsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	removeSessionFilesReferencingWorkspace(paperRoot, paths)
}

func removeSessionFilesReferencingWorkspace(paperRoot string, paths []string) {
	absRoot, err := filepath.Abs(paperRoot)
	if err != nil {
		absRoot = paperRoot
	}
	for _, path := range paths {
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
