package topicdiscovery

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var arisPromptTooLongRe = regexp.MustCompile(`prompt is too long:\s*(\d+)\s*tokens`)

type claudeProjectDirSnapshot struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	TotalBytes int64  `json:"total_bytes"`
	JSONLFiles int    `json:"jsonl_files"`
}

type arisSessionCleanupReport struct {
	AtUTC                     string                     `json:"at_utc"`
	Home                      string                     `json:"home"`
	PaperRoot                 string                     `json:"paper_root"`
	MatchTokens               []string                   `json:"match_tokens"`
	ClaudeProjectDirsRemoved  int                        `json:"claude_project_dirs_removed"`
	RemovedProjectDirNames    []string                   `json:"removed_project_dir_names,omitempty"`
	RemainingMatchingProjects []claudeProjectDirSnapshot `json:"remaining_matching_projects,omitempty"`
	NukeAllCLISessions        bool                       `json:"nuke_all_cli_sessions"`
}

type arisPromptTooLongForensic struct {
	AtUTC                string                     `json:"at_utc"`
	Home                 string                     `json:"home"`
	PaperRoot            string                     `json:"paper_root"`
	CLIPromptBytes       int                        `json:"cli_prompt_bytes"`
	PlatformPromptBytes  int                        `json:"platform_user_prompt_bytes"`
	APIReportedTokens    int                        `json:"api_reported_prompt_tokens,omitempty"`
	ErrorExcerpt         string                     `json:"error_excerpt"`
	LargestPaperFiles    []workspaceFileSize        `json:"largest_paper_files_over_64kb"`
	MatchingProjects     []claudeProjectDirSnapshot `json:"claude_projects_matching_workspace"`
	LargestProjectDirs   []claudeProjectDirSnapshot `json:"largest_claude_project_dirs_top5"`
	GlobalSessionsJSON   int                        `json:"global_claude_session_json_count"`
	GlobalSessionsBytes  int64                      `json:"global_claude_session_json_bytes"`
}

func writeARISSessionCleanupReport(paperRoot string, removed int, removedNames []string) {
	if !arisCLIWorkspaceLogEnabled() {
		return
	}
	paperRoot = filepath.Clean(strings.TrimSpace(paperRoot))
	if paperRoot == "" {
		return
	}
	home, _ := os.UserHomeDir()
	report := arisSessionCleanupReport{
		AtUTC:                    time.Now().UTC().Format(time.RFC3339),
		Home:                     home,
		PaperRoot:                paperRoot,
		MatchTokens:              claudeProjectDirMatchTokens(paperRoot),
		ClaudeProjectDirsRemoved: removed,
		RemovedProjectDirNames:   append([]string(nil), removedNames...),
		RemainingMatchingProjects: scanClaudeProjectsMatchingPaperRoot(paperRoot),
		NukeAllCLISessions:       cliSessionNukeAllEnabled(),
	}
	writeInvokeLogsJSON(paperRoot, "latest_session_cleanup.json", report)
}

func writeARISPromptTooLongForensic(paperRoot, cliPrompt, errMsg string) {
	if !arisCLIWorkspaceLogEnabled() || !strings.Contains(strings.ToLower(errMsg), "prompt is too long") {
		return
	}
	paperRoot = filepath.Clean(strings.TrimSpace(paperRoot))
	if paperRoot == "" {
		return
	}
	home, _ := os.UserHomeDir()
	platformBytes := int64(0)
	if b, err := os.ReadFile(filepath.Join(mustIdeaDiscoveryConfig().InvokeDirAbs(paperRoot), "platform_user_prompt.txt")); err == nil {
		platformBytes = int64(len(b))
	}
	tokens := 0
	if m := arisPromptTooLongRe.FindStringSubmatch(errMsg); len(m) == 2 {
		fmt.Sscanf(m[1], "%d", &tokens)
	}
	excerpt := strings.TrimSpace(errMsg)
	if len(excerpt) > 2000 {
		excerpt = excerpt[:2000]
	}
	f := arisPromptTooLongForensic{
		AtUTC:               time.Now().UTC().Format(time.RFC3339),
		Home:                home,
		PaperRoot:           paperRoot,
		CLIPromptBytes:      len(strings.TrimSpace(cliPrompt)),
		PlatformPromptBytes: int(platformBytes),
		APIReportedTokens:   tokens,
		ErrorExcerpt:        excerpt,
		LargestPaperFiles:   scanWorkspaceLargestFiles(paperRoot, 15, 64*1024),
		MatchingProjects:    scanClaudeProjectsMatchingPaperRoot(paperRoot),
		LargestProjectDirs:  largestClaudeProjectDirs(5),
	}
	sessDir := expandConfigPath("~/.claude/sessions")
	n, bytes := countSessionJSONUnder(sessDir)
	f.GlobalSessionsJSON = n
	f.GlobalSessionsBytes = bytes

	writeInvokeLogsJSON(paperRoot, "latest_prompt_too_long_forensic.json", f)
	stamp := time.Now().UTC().Format("20060102T150405.000Z")
	writeInvokeLogsJSON(paperRoot, stamp+"_prompt_too_long_forensic.json", f)
	appendARISInvokeErrorLog(paperRoot, fmt.Errorf("prompt too long forensic: home=%s matching_projects=%d largest_top=%d global_sessions=%d bytes=%d api_tokens=%d cli_prompt_bytes=%d",
		home, len(f.MatchingProjects), len(f.LargestProjectDirs), f.GlobalSessionsJSON, f.GlobalSessionsBytes, f.APIReportedTokens, f.CLIPromptBytes), 0)
}

func writeInvokeLogsJSON(paperRoot, name string, v any) {
	dir := arisInvokeLogsDir(paperRoot)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(dir, name), b, 0o644)
}

func scanClaudeProjectsMatchingPaperRoot(paperRoot string) []claudeProjectDirSnapshot {
	projectsRoot := expandConfigPath("~/.claude/projects")
	tokens := claudeProjectDirMatchTokens(paperRoot)
	return scanClaudeProjectsUnder(projectsRoot, func(name string) bool {
		return claudeProjectDirNameMatches(name, tokens)
	})
}

func largestClaudeProjectDirs(limit int) []claudeProjectDirSnapshot {
	if limit <= 0 {
		limit = 5
	}
	all := scanClaudeProjectsUnder(expandConfigPath("~/.claude/projects"), func(string) bool { return true })
	// sort by TotalBytes desc (simple bubble for tiny n)
	for i := 0; i < len(all); i++ {
		for j := i + 1; j < len(all); j++ {
			if all[j].TotalBytes > all[i].TotalBytes {
				all[i], all[j] = all[j], all[i]
			}
		}
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all
}

func scanClaudeProjectsUnder(projectsRoot string, match func(name string) bool) []claudeProjectDirSnapshot {
	st, err := os.Stat(projectsRoot)
	if err != nil || !st.IsDir() {
		return nil
	}
	entries, err := os.ReadDir(projectsRoot)
	if err != nil {
		return nil
	}
	var out []claudeProjectDirSnapshot
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		name := ent.Name()
		if match != nil && !match(name) {
			continue
		}
		path := filepath.Join(projectsRoot, name)
		snap := claudeProjectDirSnapshot{Name: name, Path: path}
		_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() {
				return nil
			}
			if strings.HasSuffix(strings.ToLower(d.Name()), ".jsonl") {
				snap.JSONLFiles++
			}
			if info, statErr := d.Info(); statErr == nil {
				snap.TotalBytes += info.Size()
			}
			return nil
		})
		out = append(out, snap)
	}
	return out
}

func countSessionJSONUnder(root string) (files int, totalBytes int64) {
	root = strings.TrimSpace(root)
	if root == "" {
		return 0, 0
	}
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		if !strings.HasSuffix(name, ".json") {
			return nil
		}
		files++
		if info, statErr := d.Info(); statErr == nil {
			totalBytes += info.Size()
		}
		return nil
	})
	return files, totalBytes
}
