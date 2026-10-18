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
		nukeAllGlobalCLISessionStores(cleanup, paperRoot)
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

func nukeAllGlobalCLISessionStores(cleanup IdeaDiscoveryCLISessionCleanup, paperRoot string) {
	dirs := []string{
		expandConfigPath(cleanup.Claude),
		expandConfigPath(cleanup.Codex),
		expandConfigPath("~/.config/aris/sessions"),
		expandConfigPath("~/.local/share/aris/sessions"),
	}
	var removed int
	for _, dir := range dirs {
		removed += nukeAllSessionJSONUnder(dir)
	}
	removed += nukeClaudeProjectsForPaperRoot(paperRoot)
	if removed > 0 {
		logger.InfoZ(context.Background(), "aris-cli-sessions-nuked",
			zap.Int("files_removed", removed),
			zap.String("paper_root", paperRoot))
	}
}

// nukeClaudeProjectsForPaperRoot 删除 ~/.claude/projects 下对应该 paper 工作区的目录（内含 *.jsonl 对话，删 inout 不会清这里）。
func nukeClaudeProjectsForPaperRoot(paperRoot string) int {
	projectsRoot := expandConfigPath("~/.claude/projects")
	st, err := os.Stat(projectsRoot)
	if err != nil || !st.IsDir() {
		return 0
	}
	slug := claudeProjectDirMatchSlug(paperRoot)
	if slug == "" {
		return 0
	}
	entries, err := os.ReadDir(projectsRoot)
	if err != nil {
		return 0
	}
	var n int
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		name := ent.Name()
		if strings.Contains(name, slug) {
			if rmErr := os.RemoveAll(filepath.Join(projectsRoot, name)); rmErr == nil {
				n++
			}
		}
	}
	return n
}

// claudeProjectDirMatchSlug 与 Claude Code 对 cwd 的目录编码一致：…/inout/user_1/…/paper → inout-user-1-manuscript-1-run-1-paper
func claudeProjectDirMatchSlug(paperRoot string) string {
	slash := filepath.ToSlash(filepath.Clean(strings.TrimSpace(paperRoot)))
	if i := strings.Index(slash, "/inout/"); i >= 0 {
		slash = slash[i+len("/inout/"):]
	}
	slash = strings.TrimPrefix(slash, "/")
	if slash == "" {
		return ""
	}
	s := strings.ReplaceAll(slash, "_", "-")
	s = strings.ReplaceAll(s, "/", "-")
	return s
}

// nukeAllSessionJSONUnder 删除目录树下全部 .json（aris/claude 会话文件命名不统一）。
func nukeAllSessionJSONUnder(root string) int {
	return nukeSessionFilesUnder(root, true)
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

// nukeSessionFilesUnder 删除目录树下 session 相关 json；allJSON 为 true 时删所有 .json/.jsonl。
func nukeSessionFilesUnder(root string, allJSON bool) int {
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
		if allJSON {
			if strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".jsonl") {
				if rmErr := os.Remove(path); rmErr == nil {
					n++
				}
			}
			return nil
		}
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
