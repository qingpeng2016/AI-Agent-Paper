package paper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
)

type UserLiteratureService struct {
	repo       repository.PaperUserLiteratureRepo
	categories repository.PaperUserLiteratureCategoryRepo
}

func NewUserLiteratureService(
	repo repository.PaperUserLiteratureRepo,
	categories repository.PaperUserLiteratureCategoryRepo,
) *UserLiteratureService {
	return &UserLiteratureService{repo: repo, categories: categories}
}

func (s *UserLiteratureService) ListMinePaged(
	ctx context.Context,
	userID uint,
	page, pageSize int,
	filter repository.UserLiteratureListFilter,
) (*response.PaperUserLiteratureListView, error) {
	page, pageSize, offset := NormalizeListPage(page, pageSize)
	uid := uint64(userID)
	total, err := s.repo.CountActiveByUser(ctx, uid, filter)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListActiveByUserPaged(ctx, uid, filter, offset, pageSize)
	if err != nil {
		return nil, err
	}
	litIDs := make([]uint64, 0, len(rows))
	for _, row := range rows {
		if row.ID > 0 {
			litIDs = append(litIDs, row.ID)
		}
	}
	catByLit := map[uint64][]repository.UserLiteratureCategoryBinding{}
	if len(litIDs) > 0 {
		bindings, err := s.repo.ListActiveCategoryBindingsByLiteratureIDs(ctx, uid, litIDs)
		if err != nil {
			return nil, err
		}
		for _, b := range bindings {
			catByLit[b.UserLiteratureID] = append(catByLit[b.UserLiteratureID], b)
		}
	}
	out := &response.PaperUserLiteratureListView{
		Items:    make([]response.PaperUserLiteratureItemView, 0, len(rows)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, row := range rows {
		lit := row.PaperUserLiterature
		item := response.PaperUserLiteratureItemView{
			ID:              strconv.FormatUint(lit.ID, 10),
			From:            lit.From,
			FileExt:         lit.FileExt,
			FileName:        lit.FileName,
			LiteratureTitle: lit.LiteratureTitle,
			SizeBytes:       lit.SizeBytes,
		}
		for _, b := range catByLit[lit.ID] {
			item.CategoryIDs = append(item.CategoryIDs, strconv.FormatUint(b.CategoryID, 10))
			name := strings.TrimSpace(b.CategoryName)
			if name != "" {
				item.CategoryNames = append(item.CategoryNames, name)
			}
		}
		if lit.Note != nil {
			item.Note = *lit.Note
		}
		if lit.StorageURI != nil {
			item.StorageURI = *lit.StorageURI
		}
		if !lit.CreatedAt.IsZero() {
			item.CreatedAt = lit.CreatedAt.UTC().Format(time.RFC3339)
		}
		out.Items = append(out.Items, item)
	}
	return out, nil
}

func (s *UserLiteratureService) ListCategoriesPaged(
	ctx context.Context,
	page, pageSize int,
) (*response.PaperUserLiteratureCategoryListView, error) {
	page, pageSize, offset := NormalizeListPage(page, pageSize)
	total, err := s.categories.CountActive(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.categories.ListActivePaged(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}
	out := &response.PaperUserLiteratureCategoryListView{
		Items:    make([]response.PaperUserLiteratureCategoryItemView, 0, len(rows)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, row := range rows {
		item := response.PaperUserLiteratureCategoryItemView{
			ID:     strconv.FormatUint(row.ID, 10),
			Name:   row.Name,
			Sort:   row.Sort,
			Status: row.Status,
		}
		if !row.CreatedAt.IsZero() {
			item.CreatedAt = row.CreatedAt.UTC().Format(time.RFC3339)
		}
		out.Items = append(out.Items, item)
	}
	return out, nil
}
