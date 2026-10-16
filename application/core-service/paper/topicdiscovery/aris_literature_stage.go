package topicdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/literature"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
)

const arisLiteratureSubdir = "literature"

type arisLiteratureStagingEntry struct {
	ExternalKey     string `json:"external_key"`
	PlatformRelPath string `json:"platform_rel_path"`
	WorkspacePath   string `json:"workspace_path"`
	Bytes           int64  `json:"bytes,omitempty"`
}

// enrichCorpusWithRetrieveExtra 把 retrieve.extra 里的 literature_downloads 并入 corpus，供 invoke 落盘与 staging。
func enrichCorpusWithRetrieveExtra(corpus json.RawMessage, retrieveExtra []byte) json.RawMessage {
	downloads := literature.ParseLiteratureDownloads(retrieveExtra)
	if len(downloads) == 0 {
		return corpus
	}
	var m map[string]any
	if len(corpus) > 0 && json.Unmarshal(corpus, &m) != nil {
		m = nil
	}
	if m == nil {
		m = map[string]any{}
	}
	m["literature_downloads"] = downloads
	b, err := json.Marshal(m)
	if err != nil {
		return corpus
	}
	return json.RawMessage(b)
}

func parseLiteratureDownloadsFromCorpus(corpus json.RawMessage) []literature.LiteratureDownloadItem {
	if len(corpus) == 0 {
		return nil
	}
	var m map[string]any
	if json.Unmarshal(corpus, &m) != nil {
		return nil
	}
	raw, ok := m["literature_downloads"]
	if !ok {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var items []literature.LiteratureDownloadItem
	if json.Unmarshal(b, &items) != nil {
		return nil
	}
	return items
}

// SyncRunLiteratureToInoutWorkspace 将 retrieve 下载 + 用户「我的文献」PDF 复制到 inout/.../paper/literature/。
func (s *TopicDiscoveryRunService) SyncRunLiteratureToInoutWorkspace(
	ctx context.Context,
	userID uint,
	runVersion int,
	userLitRepo repository.PaperUserLiteratureRepo,
) (int, error) {
	if s == nil || s.steps == nil {
		return 0, nil
	}
	ret, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "retrieve")
	if err != nil || ret == nil {
		return 0, err
	}
	items, err := s.collectRunLiteratureItems(ctx, userID, runVersion, userLitRepo)
	if err != nil {
		return 0, err
	}
	paperRoot := ARISRunPaperRoot(uint64(userID), ret.ManuscriptID, runVersion)
	return syncLiteratureItemsToPaperRoot(paperRoot, items)
}

func syncLiteratureItemsToPaperRoot(paperRoot string, items []literature.LiteratureDownloadItem) (int, error) {
	if len(items) == 0 {
		return 0, nil
	}
	corpus, err := json.Marshal(map[string]any{"literature_downloads": items})
	if err != nil {
		return 0, err
	}
	staged, err := stageRetrieveLiteraturePDFs(paperRoot, json.RawMessage(corpus))
	if err != nil {
		return 0, err
	}
	if err := writeLiteratureStagingManifest(paperRoot, staged); err != nil {
		return 0, err
	}
	return len(staged), nil
}

func (s *TopicDiscoveryRunService) collectRunLiteratureItems(
	ctx context.Context,
	userID uint,
	runVersion int,
	userLitRepo repository.PaperUserLiteratureRepo,
) ([]literature.LiteratureDownloadItem, error) {
	ret, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "retrieve")
	if err != nil || ret == nil {
		return nil, err
	}
	items := literature.ParseLiteratureDownloads(ret.Result)
	items = appendUserLiteratureDownloadItems(ctx, s, userLitRepo, userID, runVersion, items)
	return dedupeLiteratureDownloadItems(items), nil
}

func appendUserLiteratureDownloadItems(
	ctx context.Context,
	s *TopicDiscoveryRunService,
	userLitRepo repository.PaperUserLiteratureRepo,
	userID uint,
	runVersion int,
	base []literature.LiteratureDownloadItem,
) []literature.LiteratureDownloadItem {
	if userLitRepo == nil || s == nil {
		return base
	}
	input, err := s.loadRunInput(ctx, userID, runVersion)
	if err != nil || len(input.UserLiteratureIDs) == 0 {
		return base
	}
	want := make(map[uint64]struct{}, len(input.UserLiteratureIDs))
	for _, id := range input.UserLiteratureIDs {
		if id > 0 {
			want[id] = struct{}{}
		}
	}
	rows, err := userLitRepo.ListActiveByUser(ctx, uint64(userID))
	if err != nil {
		return base
	}
	out := append([]literature.LiteratureDownloadItem(nil), base...)
	for _, row := range rows {
		if _, ok := want[row.ID]; !ok {
			continue
		}
		if row.FileExt != "pdf" && !strings.HasSuffix(strings.ToLower(row.FileName), ".pdf") {
			continue
		}
		rel := ""
		if row.StorageURI != nil {
			rel = strings.TrimSpace(*row.StorageURI)
		}
		if rel == "" {
			continue
		}
		out = append(out, literature.LiteratureDownloadItem{
			ExternalKey: fmt.Sprintf("user-lit-%d", row.ID),
			LocalPath:   filepath.ToSlash(rel),
		})
	}
	return out
}

func dedupeLiteratureDownloadItems(items []literature.LiteratureDownloadItem) []literature.LiteratureDownloadItem {
	if len(items) == 0 {
		return items
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]literature.LiteratureDownloadItem, 0, len(items))
	for _, it := range items {
		key := strings.TrimSpace(it.LocalPath)
		if key == "" {
			key = strings.TrimSpace(it.ExternalKey)
		}
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, it)
	}
	return out
}

// stageRetrieveStepLiteratureToWorkspace 把 retrieve 步已落盘 PDF 同步到本 run 的 paperRoot/literature/。
func stageRetrieveStepLiteratureToWorkspace(paperRoot string, s *TopicDiscoveryRunService, ctx context.Context, userID uint, runVersion int) error {
	if s == nil || s.steps == nil {
		return nil
	}
	items, err := s.collectRunLiteratureItems(ctx, userID, runVersion, nil)
	if err != nil {
		return err
	}
	_, err = syncLiteratureItemsToPaperRoot(paperRoot, items)
	return err
}

func literatureStagingFileName(externalKey, platformRel string) string {
	if ek := strings.TrimSpace(externalKey); ek != "" {
		base := sanitizePathSegment(ek)
		if base != "" {
			if !strings.HasSuffix(strings.ToLower(base), ".pdf") {
				base += ".pdf"
			}
			return base
		}
	}
	name := filepath.Base(platformRel)
	if name == "" || name == "." {
		return "literature.pdf"
	}
	return name
}

// stageRetrieveLiteraturePDFs 将平台 storage 中的 PDF 放入 ARIS workspace 的 literature/（research-lit 默认识别路径）。
func stageRetrieveLiteraturePDFs(paperRoot string, corpus json.RawMessage) ([]arisLiteratureStagingEntry, error) {
	items := parseLiteratureDownloadsFromCorpus(corpus)
	if len(items) == 0 {
		return nil, nil
	}
	litDir := filepath.Join(paperRoot, arisLiteratureSubdir)
	if err := os.MkdirAll(litDir, 0o755); err != nil {
		return nil, err
	}
	var staged []arisLiteratureStagingEntry
	for _, it := range items {
		rel := strings.TrimSpace(it.LocalPath)
		if rel == "" {
			continue
		}
		src := literature.LiteratureLocalAbsPath(rel)
		st, err := os.Stat(src)
		if err != nil || st.Size() == 0 {
			continue
		}
		name := literatureStagingFileName(it.ExternalKey, rel)
		dst := filepath.Join(litDir, name)
		if err := linkOrCopyPDF(src, dst); err != nil {
			return staged, fmt.Errorf("stage pdf %s: %w", it.ExternalKey, err)
		}
		wsRel := filepath.ToSlash(filepath.Join(arisLiteratureSubdir, name))
		staged = append(staged, arisLiteratureStagingEntry{
			ExternalKey:     strings.TrimSpace(it.ExternalKey),
			PlatformRelPath: filepath.ToSlash(rel),
			WorkspacePath:   wsRel,
			Bytes:           st.Size(),
		})
	}
	return staged, nil
}

func linkOrCopyPDF(src, dst string) error {
	if st, err := os.Stat(dst); err == nil && st.Size() > 0 {
		return nil
	}
	_ = os.Remove(dst)
	if err := os.Symlink(src, dst); err == nil {
		return nil
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(dst)
		return copyErr
	}
	return closeErr
}

func writeLiteratureStagingManifest(paperRoot string, staged []arisLiteratureStagingEntry) error {
	if len(staged) == 0 {
		return nil
	}
	dir := filepath.Join(paperRoot, ".paper")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	payload := map[string]any{
		"note":    "Platform retrieve PDFs staged for /research-lit local library (literature/*.pdf).",
		"entries": staged,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "literature_staging.json"), b, 0o644)
}
