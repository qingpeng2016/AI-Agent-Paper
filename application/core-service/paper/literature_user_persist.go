package paper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
)

func literatureHitsIndex(resultJSON []byte) map[string]storedLiteratureHit {
	out := map[string]storedLiteratureHit{}
	if len(resultJSON) == 0 {
		return out
	}
	var wrap struct {
		Hits []storedLiteratureHit `json:"literature_hits"`
	}
	if json.Unmarshal(resultJSON, &wrap) != nil {
		return out
	}
	for _, h := range wrap.Hits {
		key := strings.TrimSpace(h.ExternalKey)
		if key == "" {
			continue
		}
		out[key] = h
	}
	return out
}

func userLiteratureFromPlatform(code, externalKey string) string {
	code = strings.TrimSpace(code)
	if code != "" {
		return code
	}
	ext := strings.TrimSpace(externalKey)
	switch {
	case strings.HasPrefix(ext, "arxiv:"):
		return entity.UserLiteratureFromArxiv
	case strings.HasPrefix(ext, "openalex:"):
		return entity.UserLiteratureFromOpenAlex
	case strings.HasPrefix(ext, "s2:"):
		return entity.UserLiteratureFromSemanticScholar
	case strings.HasPrefix(ext, "crossref:"):
		return entity.UserLiteratureFromCrossref
	case strings.HasPrefix(ext, "pubmed:"):
		return entity.UserLiteratureFromPubMed
	default:
		return entity.UserLiteratureFromArxiv
	}
}

func sha256FileHex(absPath string) (string, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// PersistDownloadedUserLiterature 将已落盘的 retrieve PDF 写入 paper_user_literature（按 user+storage_uri 去重）。
func PersistDownloadedUserLiterature(
	ctx context.Context,
	repo repository.PaperUserLiteratureRepo,
	userID uint64,
	resultJSON []byte,
	item LiteratureDownloadItem,
) error {
	if repo == nil || userID == 0 {
		return nil
	}
	rel := strings.TrimSpace(item.LocalPath)
	if rel == "" {
		rel = LiteratureLocalRelPath(item.ExternalKey)
	}
	abs := LiteratureLocalAbsPath(rel)
	st, err := os.Stat(abs)
	if err != nil || st.Size() == 0 {
		return nil
	}

	existing, err := repo.FindActiveByUserStorageURI(ctx, userID, rel)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	hit := literatureHitsIndex(resultJSON)[strings.TrimSpace(item.ExternalKey)]
	from := userLiteratureFromPlatform(hit.SourceCode, item.ExternalKey)
	title := strings.TrimSpace(hit.Title)
	if title == "" {
		title = strings.TrimSpace(item.ExternalKey)
	}
	if title == "" {
		title = "文献"
	}
	fileName := literatureLocalPDFFileName(item.ExternalKey)
	if strings.TrimSpace(fileName) == "" {
		fileName = "literature.pdf"
	}

	var hashPtr *string
	if sum, err := sha256FileHex(abs); err == nil && sum != "" {
		hashPtr = &sum
	}
	storageURI := rel
	size := uint64(st.Size())

	row := &entity.PaperUserLiterature{
		UserID:          userID,
		From:            from,
		FileExt:         "pdf",
		FileName:        fileName,
		LiteratureTitle: title,
		SizeBytes:       size,
		StorageURI:      &storageURI,
		ContentHash:     hashPtr,
		Status:          "active",
	}
	return repo.Create(ctx, row)
}

// SyncDownloadedUserLiterature 批量同步已成功落盘的下载项。
func SyncDownloadedUserLiterature(
	ctx context.Context,
	repo repository.PaperUserLiteratureRepo,
	userID uint64,
	resultJSON []byte,
	items []LiteratureDownloadItem,
) []error {
	var errs []error
	for _, it := range items {
		if strings.TrimSpace(it.PdfURL) == "" {
			continue
		}
		if err := PersistDownloadedUserLiterature(ctx, repo, userID, resultJSON, it); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
