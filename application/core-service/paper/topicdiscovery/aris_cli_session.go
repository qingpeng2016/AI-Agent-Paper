package topicdiscovery

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"go.uber.org/zap"
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
	clearWorkspaceLocalSessionFiles(paperRoot)
	if cliSessionNukeAllEnabled() {
		nukeAllGlobalCLISessionStores(cleanup)
	} else {
		clearGlobalSessionsClaude(paperRoot, expandConfigPath(cleanup.Claude))
		clearGlobalSessionsCodex(paperRoot, expandConfigPath(cleanup.Codex))
		clearGlobalSessionsWalkDir(paperRoot, expandConfigPath("~/.config/aris/sessions"))
	}
}

// cliSessionNukeAllEnabled 为 true 时清空 bot 用户下全部 CLI session 文件（删 inout 无法触及 ~/.claude 等）。
// 设 PAPER_ARIS_NUKE_ALL_CLI_SESSIONS=0 可关闭（同机多任务并行 invoke 时慎用）。
func cliSessionNukeAllEnabled() bool {
	v := strings.TrimSpace(os.Getenv("PAPER_ARIS_NUKE_ALL_CLI_SESSIONS"))
	if v == "" {
		return true
	}
	switch strings.ToLower(v) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func nukeAllGlobalCLISessionStores(cleanup IdeaDiscoveryCLISessionCleanup) {
	dirs := []string{
		expandConfigPath(cleanup.Claude),
		expandConfigPath(cleanup.Codex),
		expandConfigPath("~/.config/aris/sessions"),
		expandConfigPath("~/.local/share/aris/sessions"),
	}
	var removed int
	for _, dir := range dirs {
		removed += nukeSessionFilesUnder(dir)
	}
	if removed > 0 {
		logger.InfoZ(context.Background(), "aris-cli-sessions-nuked", zap.Int("files_removed", removed))
	}
}

func clearWorkspaceLocalSessionFiles(paperRoot string) {
	for _, name := range []string{"session.json", ".session.json", "aris-session.json"} {
		_ = os.Remove(filepath.Join(paperRoot, name))
	}
	arisDir := filepath.Join(paperRoot, ".aris")
	if st, err := os.Stat(arisDir); err == nil && st.IsDir() {
		_ = filepath.WalkDir(arisDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			name := strings.ToLower(d.Name())
			if strings.HasSuffix(name, ".json") && strings.Contains(name, "session") {
				_ = os.Remove(path)
			}
			return nil
		})
	}
}

// nukeSessionFilesUnder 删除目录树下所有疑似 session 的 json（不读 cwd 匹配，避免 path 写法不一致导致删不掉）。
func nukeSessionFilesUnder(root string) int {
	root = strings.TrimSpace(root)
	if root == "" {
		return 0
	}
	st, err := os.Stat(root)
	if err != nil || !st.IsDir() {
		return 0
	}
	var n int
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		if !strings.HasSuffix(name, ".json") {
			return nil
		}
		if strings.HasPrefix(name, "session-") || strings.Contains(name, "session") {
			if rmErr := os.Remove(path); rmErr == nil {
				n++
			}
		}
		return nil
	})
	return n
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
	clearGlobalSessionsWalkDir(paperRoot, sessionsDir)
}

func clearGlobalSessionsWalkDir(paperRoot, sessionsDir string) {
	if sessionsDir == "" {
		return
	}
	if st, err := os.Stat(sessionsDir); err != nil || !st.IsDir() {
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
	needles := workspaceSessionMatchNeedles(paperRoot)
	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		text := string(b)
		if sessionTextReferencesWorkspace(text, needles) {
			_ = os.Remove(path)
		}
	}
}

// workspaceSessionMatchNeedles 全局 session JSON 里 cwd/path 写法不一（绝对路径、inout/…、user_N/…），多片段匹配提高删除率。
func workspaceSessionMatchNeedles(paperRoot string) []string {
	paperRoot = filepath.Clean(strings.TrimSpace(paperRoot))
	absRoot, err := filepath.Abs(paperRoot)
	if err != nil {
		absRoot = paperRoot
	}
	seen := map[string]struct{}{}
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
	}
	for _, p := range []string{paperRoot, absRoot} {
		add(p)
		add(filepath.ToSlash(p))
	}
	slash := filepath.ToSlash(absRoot)
	if i := strings.Index(slash, "/inout/"); i >= 0 {
		add(slash[i+1:])
		add(slash[i+len("/inout/"):])
	}
	parts := strings.Split(filepath.ToSlash(absRoot), "/")
	for i := 0; i < len(parts); i++ {
		if strings.HasPrefix(parts[i], "user_") {
			add(strings.Join(parts[i:], "/"))
			break
		}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	return out
}

func sessionTextReferencesWorkspace(text string, needles []string) bool {
	for _, n := range needles {
		if n != "" && strings.Contains(text, n) {
			return true
		}
	}
	return false
}
