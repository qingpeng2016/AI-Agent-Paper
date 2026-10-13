package paper

import (
	"context"
	"strconv"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
)

type UserLiteratureService struct {
	repo repository.PaperUserLiteratureRepo
}

func NewUserLiteratureService(repo repository.PaperUserLiteratureRepo) *UserLiteratureService {
	return &UserLiteratureService{repo: repo}
}

func (s *UserLiteratureService) ListMinePaged(
	ctx context.Context,
	userID uint,
	page, pageSize int,
) (*response.PaperUserLiteratureListView, error) {
	page, pageSize, offset := NormalizeListPage(page, pageSize)
	uid := uint64(userID)
	total, err := s.repo.CountActiveByUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListActiveByUserPaged(ctx, uid, offset, pageSize)
	if err != nil {
		return nil, err
	}
	out := &response.PaperUserLiteratureListView{
		Items:    make([]response.PaperUserLiteratureItemView, 0, len(rows)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, row := range rows {
		item := response.PaperUserLiteratureItemView{
			ID:              strconv.FormatUint(row.ID, 10),
			From:            row.From,
			FileExt:         row.FileExt,
			FileName:        row.FileName,
			LiteratureTitle: row.LiteratureTitle,
			SizeBytes:       row.SizeBytes,
		}
		if row.Note != nil {
			item.Note = *row.Note
		}
		if row.StorageURL != nil {
			item.StorageURL = *row.StorageURL
		}
		if !row.CreatedAt.IsZero() {
			item.CreatedAt = row.CreatedAt.UTC().Format(time.RFC3339)
		}
		out.Items = append(out.Items, item)
	}
	return out, nil
}
