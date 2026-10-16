package topicdiscovery

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/literature"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

// ingestARISWorkspaceLiterature 将 paperRoot/literature/*.pdf 复制到 storage/paper-files 并写入 paper_user_literature。
func (s *TopicDiscoveryRunService) ingestARISWorkspaceLiterature(
	ctx context.Context,
	userID uint,
	runVersion int,
	manuscriptID uint64,
	paperRoot string,
	cfg ARISIdeaDiscoveryConfig,
) (int, error) {
	if s == nil || s.userLit == nil {
		return 0, nil
	}
	litRel := strings.TrimSpace(cfg.Artifacts.LiteratureDir)
	if litRel == "" {
		litRel = arisLiteratureSubdir
	}
	litAbs := filepath.Join(paperRoot, filepath.FromSlash(litRel))
	entries, err := os.ReadDir(litAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	corpusMeta := readCorpusPaperMeta(filepath.Join(paperRoot, filepath.FromSlash(strings.TrimSpace(cfg.Artifacts.CorpusJSON))))
	var items []literature.LiteratureDownloadItem
	var hits []storedLiteratureHit
	n := 0

	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(strings.ToLower(ent.Name()), ".pdf") {
			continue
		}
		src := filepath.Join(litAbs, ent.Name())
		st, err := os.Stat(src)
		if err != nil || st.Size() == 0 {
			continue
		}
		extKey, from := externalKeyFromWorkspacePDF(ent.Name())
		meta := corpusMeta[strings.ToLower(ent.Name())]
		title := meta.title
		if title == "" {
			title = strings.TrimSuffix(ent.Name(), filepath.Ext(ent.Name()))
		}
		rel := literature.LiteratureLocalRelPath(extKey)
		dst := literature.LiteratureLocalAbsPath(rel)
		if err := copyFile(src, dst); err != nil {
			continue
		}
		items = append(items, literature.LiteratureDownloadItem{
			ExternalKey: extKey,
			LocalPath:   rel,
			PdfURL:      meta.url,
		})
		hits = append(hits, storedLiteratureHit{
			SourceCode:  from,
			ExternalKey: extKey,
			Title:       title,
			URL:         meta.url,
			QueryText:   "aris_research_lit",
		})
		if err := literature.PersistIngestedUserLiteraturePDF(ctx, s.userLit, uint64(userID), extKey, from, title, rel); err != nil {
			continue
		}
		n++
	}

	if n == 0 {
		return 0, nil
	}
	_ = s.mergeRetrieveStepLiteratureResult(ctx, userID, runVersion, manuscriptID, hits, items, "aris_web_research_lit")
	return n, nil
}

type corpusPaperMeta struct {
	title string
	url   string
}

func readCorpusPaperMeta(corpusPath string) map[string]corpusPaperMeta {
	out := map[string]corpusPaperMeta{}
	b, err := os.ReadFile(corpusPath)
	if err != nil {
		return out
	}
	var root map[string]any
	if json.Unmarshal(b, &root) != nil {
		return out
	}
	papers, _ := root["papers"].([]any)
	for _, p := range papers {
		m, _ := p.(map[string]any)
		if m == nil {
			continue
		}
		fileName := strings.ToLower(strings.TrimSpace(stringField(m, "local_file", "file", "pdf", "filename")))
		if fileName == "" {
			continue
		}
		out[filepath.Base(fileName)] = corpusPaperMeta{
			title: stringField(m, "title", "name"),
			url:   stringField(m, "url", "link", "arxiv_url"),
		}
	}
	return out
}

func stringField(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			s := strings.TrimSpace(strings.Trim(fmtAny(v), `"`))
			s = strings.Trim(s, `"`)
			if s != "" {
				return s
			}
		}
	}
	return ""
}

func fmtAny(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

func externalKeyFromWorkspacePDF(fileName string) (externalKey, from string) {
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	base = strings.TrimSpace(base)
	lower := strings.ToLower(base)
	if strings.HasPrefix(lower, "arxiv_") {
		id := strings.ReplaceAll(base[len("arxiv_"):], "_", ".")
		return "arxiv:" + id, entity.UserLiteratureFromArxiv
	}
	if strings.Contains(base, ":") {
		parts := strings.SplitN(base, ":", 2)
		return strings.TrimSpace(base), userLiteratureFromCode(parts[0])
	}
	return "aris:" + sanitizePathSegment(base), entity.UserLiteratureFromArxiv
}

func userLiteratureFromCode(code string) string {
	code = strings.TrimSpace(strings.ToLower(code))
	switch code {
	case "arxiv":
		return entity.UserLiteratureFromArxiv
	case "openalex":
		return entity.UserLiteratureFromOpenAlex
	default:
		return entity.UserLiteratureFromArxiv
	}
}

func (s *TopicDiscoveryRunService) mergeRetrieveStepLiteratureResult(
	ctx context.Context,
	userID uint,
	runVersion int,
	manuscriptID uint64,
	hits []storedLiteratureHit,
	items []literature.LiteratureDownloadItem,
	source string,
) error {
	if s == nil || s.steps == nil {
		return nil
	}
	step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "retrieve")
	if err != nil || step == nil {
		return err
	}
	var prev map[string]any
	_ = json.Unmarshal(step.Result, &prev)
	if prev == nil {
		prev = map[string]any{}
	}
	prev["skipped"] = true
	prev["reason"] = "native_aris_web_literature"
	prev["literature_source"] = source
	prev["hit_count"] = len(hits)
	prev["verified_count"] = len(hits)
	prev["literature_hits"] = hits
	prev["literature_downloads"] = items
	if _, ok := prev["search_query"]; !ok {
		prev["search_query"] = ""
	}
	step.Result = jsonx.MustJSON(prev)
	step.ExecProgressPct = 100
	if strings.TrimSpace(step.Status) != "completed" {
		step.Status = "completed"
	}
	return s.steps.SaveStep(ctx, step)
}

func isResearchLitPhaseDone(paperRoot, runID string) bool {
	cfg := mustIdeaDiscoveryConfig()
	b, err := os.ReadFile(cfg.RunStateFileAbs(paperRoot, runID))
	if err != nil {
		return false
	}
	var st struct {
		Phases []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"phases"`
	}
	if json.Unmarshal(b, &st) != nil {
		return false
	}
	for _, p := range st.Phases {
		id := strings.TrimSpace(p.ID)
		if id == "" {
			id = strings.TrimSpace(p.Name)
		}
		if id != "research-lit" {
			continue
		}
		switch strings.TrimSpace(strings.ToLower(p.Status)) {
		case "done", "accepted", "provisional", "completed":
			return true
		default:
			return false
		}
	}
	return false
}
