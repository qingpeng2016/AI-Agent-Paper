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

// ClearResearchLitInvokeWorkspace 在 /research-lit 前清理会触发 ARIS 把整库塞进首包上下文的残留
//（同 cwd 下即使用 "ping" 也会爆 token；空目录则正常）。保留 RESEARCH_BRIEF、.paper/invoke 审计日志。
func (c ARISIdeaDiscoveryConfig) ClearResearchLitInvokeWorkspace(paperRoot string) {
	paperRoot = filepath.Clean(strings.TrimSpace(paperRoot))
	if paperRoot == "" {
		return
	}
	var removedPaths []string
	addRemoved := func(rel string) {
		rel = strings.TrimSpace(filepath.ToSlash(rel))
		if rel != "" {
			removedPaths = append(removedPaths, rel)
		}
	}

	litDir := strings.TrimSpace(c.Artifacts.LiteratureDir)
	if litDir == "" {
		litDir = "literature"
	}
	if rmDir(paperRoot, litDir) {
		addRemoved(litDir + "/")
	}

	for _, rel := range []string{
		c.Artifacts.CorpusJSON,
		".paper/literature_staging.json",
		"LITERATURE_LANDSCAPE.md",
		"LITERATURE_REVIEW.md",
		"LITERATURE_REVIEW_SUMMARY.txt",
		"EXECUTIVE_SUMMARY.md",
		"references.bib",
	} {
		rel = strings.TrimSpace(filepath.ToSlash(rel))
		if rel == "" {
			continue
		}
		if rmPath(filepath.Join(paperRoot, rel)) {
			addRemoved(rel)
		}
	}

	// 失败重跑时清掉 .paper 内除 invoke 外的缓存（corpus 等可能在 .paper/ 根下）
	paperMeta := filepath.Join(paperRoot, ".paper")
	_ = filepath.WalkDir(paperMeta, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(paperRoot, path)
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, ".paper/invoke/") {
			return nil
		}
		if rmPath(path) {
			addRemoved(rel)
		}
		return nil
	})

	large := scanWorkspaceLargestFiles(paperRoot, 15, 256*1024)
	logger.InfoZ(context.Background(), "aris-research-lit-workspace-prune",
		zap.String("paper_root", paperRoot),
		zap.Strings("removed", removedPaths),
		zap.Any("largest_files_after", large))
}

func rmDir(root, rel string) bool {
	p := filepath.Join(root, rel)
	st, err := os.Stat(p)
	if err != nil || !st.IsDir() {
		return false
	}
	return os.RemoveAll(p) == nil
}

func rmPath(p string) bool {
	if err := os.Remove(p); err == nil {
		return true
	}
	if st, err := os.Stat(p); err == nil && st.IsDir() {
		return os.RemoveAll(p) == nil
	}
	return false
}

type workspaceFileSize struct {
	Rel  string `json:"rel"`
	Bytes int64  `json:"bytes"`
}

func scanWorkspaceLargestFiles(paperRoot string, limit int, minBytes int64) []workspaceFileSize {
	if limit <= 0 {
		limit = 10
	}
	var all []workspaceFileSize
	_ = filepath.WalkDir(paperRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil || info.Size() < minBytes {
			return nil
		}
		rel, _ := filepath.Rel(paperRoot, path)
		all = append(all, workspaceFileSize{Rel: filepath.ToSlash(rel), Bytes: info.Size()})
		return nil
	})
	for i := 0; i < len(all); i++ {
		for j := i + 1; j < len(all); j++ {
			if all[j].Bytes > all[i].Bytes {
				all[i], all[j] = all[j], all[i]
			}
		}
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all
}
