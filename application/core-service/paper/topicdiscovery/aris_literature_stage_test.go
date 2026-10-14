package topicdiscovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/literature"
)

func TestStageRetrieveLiteraturePDFs(t *testing.T) {
	root := t.TempDir()
	pdfDir := filepath.Join(root, "paper-files")
	if err := os.MkdirAll(pdfDir, 0o755); err != nil {
		t.Fatal(err)
	}
	pdfPath := filepath.Join(pdfDir, "arxiv_test.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4 test"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PAPER_STORAGE_ROOT", root)

	corpus, _ := json.Marshal(map[string]any{
		"literature_hits": []any{},
		"literature_downloads": []literature.LiteratureDownloadItem{{
			ExternalKey: "arxiv:test",
			LocalPath:   "paper-files/arxiv_test.pdf",
		}},
	})
	paperRoot := filepath.Join(t.TempDir(), "paper")
	staged, err := stageRetrieveLiteraturePDFs(paperRoot, corpus)
	if err != nil {
		t.Fatal(err)
	}
	if len(staged) != 1 {
		t.Fatalf("staged %d want 1", len(staged))
	}
	ws := filepath.Join(paperRoot, staged[0].WorkspacePath)
	st, err := os.Stat(ws)
	if err != nil || st.Size() == 0 {
		t.Fatalf("workspace pdf missing: %v", err)
	}
}

func TestEnrichCorpusWithRetrieveExtra(t *testing.T) {
	extra, _ := json.Marshal(map[string]any{
		"literature_downloads": []map[string]string{
			{"external_key": "a:1", "local_path": "paper-files/x.pdf"},
		},
	})
	out := enrichCorpusWithRetrieveExtra(json.RawMessage(`{"literature_hits":[]}`), extra)
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["literature_downloads"]; !ok {
		t.Fatal("missing literature_downloads")
	}
}
