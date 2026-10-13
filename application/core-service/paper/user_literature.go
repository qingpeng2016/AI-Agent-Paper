package paper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
)

type UserLiteratureService struct {
	repo          repository.PaperUserLiteratureRepo
	categories    repository.PaperUserLiteratureCategoryRepo
	manuscripts   repository.PaperManuscriptRepo
	msLiteratureRel repository.PaperUserLiteratureManuscriptRelRepo
}

func NewUserLiteratureService(
	repo repository.PaperUserLiteratureRepo,
	categories repository.PaperUserLiteratureCategoryRepo,
	manuscripts repository.PaperManuscriptRepo,
	msLiteratureRel repository.PaperUserLiteratureManuscriptRelRepo,
) *UserLiteratureService {
	return &UserLiteratureService{
		repo:            repo,
		categories:      categories,
		manuscripts:     manuscripts,
		msLiteratureRel: msLiteratureRel,
	}
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
	msByLit := map[uint64][]repository.LiteratureManuscriptBinding{}
	if len(litIDs) > 0 {
		bindings, err := s.repo.ListActiveCategoryBindingsByLiteratureIDs(ctx, uid, litIDs)
		if err != nil {
			return nil, err
		}
		for _, b := range bindings {
			catByLit[b.UserLiteratureID] = append(catByLit[b.UserLiteratureID], b)
		}
		msBindings, err := s.msLiteratureRel.ListActiveManuscriptBindingsByLiteratureIDs(ctx, userID, litIDs)
		if err != nil {
			return nil, err
		}
		for _, b := range msBindings {
			msByLit[b.UserLiteratureID] = append(msByLit[b.UserLiteratureID], b)
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
		for _, b := range msByLit[lit.ID] {
			item.ManuscriptIDs = append(item.ManuscriptIDs, strconv.FormatUint(b.ManuscriptID, 10))
			title := strings.TrimSpace(b.ManuscriptTitle)
			if title != "" {
				item.ManuscriptTitles = append(item.ManuscriptTitles, title)
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
			Status: row.Status,
		}
		if !row.CreatedAt.IsZero() {
			item.CreatedAt = row.CreatedAt.UTC().Format(time.RFC3339)
		}
		out.Items = append(out.Items, item)
	}
	return out, nil
}

func (s *UserLiteratureService) CreateCategory(
	ctx context.Context,
	name string,
) (*response.PaperUserLiteratureCategoryItemView, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 128 {
		return nil, errorx.ErrParamsError
	}
	exists, err := s.categories.ExistsActiveByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errorx.ErrParamsError.WithDetail("分类名称已存在")
	}
	row := &entity.PaperUserLiteratureCategory{
		Name:   name,
		Status: "active",
	}
	if err := s.categories.Create(ctx, row); err != nil {
		return nil, err
	}
	item := response.PaperUserLiteratureCategoryItemView{
		ID:     strconv.FormatUint(row.ID, 10),
		Name:   row.Name,
		Status: row.Status,
	}
	if !row.CreatedAt.IsZero() {
		item.CreatedAt = row.CreatedAt.UTC().Format(time.RFC3339)
	}
	return &item, nil
}

func (s *UserLiteratureService) UpdateCategory(
	ctx context.Context,
	categoryIDRaw string,
	name string,
) (*response.PaperUserLiteratureCategoryItemView, error) {
	categoryID, err := parseUserLiteratureCategoryID(categoryIDRaw)
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 128 {
		return nil, errorx.ErrParamsError
	}
	cat, err := s.categories.FindActiveByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, errorx.ErrParamsError.WithDetail("分类不存在")
	}
	if cat.Name == name {
		return categoryItemView(cat), nil
	}
	exists, err := s.categories.ExistsActiveByNameExceptID(ctx, name, categoryID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errorx.ErrParamsError.WithDetail("分类名称已存在")
	}
	if err := s.categories.UpdateActiveName(ctx, categoryID, name); err != nil {
		return nil, err
	}
	updated, err := s.categories.FindActiveByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, errorx.ErrParamsError.WithDetail("分类不存在")
	}
	return categoryItemView(updated), nil
}

func categoryItemView(row *entity.PaperUserLiteratureCategory) *response.PaperUserLiteratureCategoryItemView {
	if row == nil {
		return nil
	}
	item := &response.PaperUserLiteratureCategoryItemView{
		ID:     strconv.FormatUint(row.ID, 10),
		Name:   row.Name,
		Status: row.Status,
	}
	if !row.CreatedAt.IsZero() {
		item.CreatedAt = row.CreatedAt.UTC().Format(time.RFC3339)
	}
	return item
}

func (s *UserLiteratureService) ListCategoryLiteratureIDs(
	ctx context.Context,
	userID uint,
	categoryIDRaw string,
) (*response.PaperUserLiteratureCategoryLiteratureIDsView, error) {
	categoryID, err := parseUserLiteratureCategoryID(categoryIDRaw)
	if err != nil {
		return nil, err
	}
	cat, err := s.categories.FindActiveByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, errorx.ErrParamsError.WithDetail("分类不存在")
	}
	uid := uint64(userID)
	ids, err := s.repo.ListActiveLiteratureIDsByCategory(ctx, uid, categoryID)
	if err != nil {
		return nil, err
	}
	out := &response.PaperUserLiteratureCategoryLiteratureIDsView{
		CategoryID:    strconv.FormatUint(categoryID, 10),
		LiteratureIDs: make([]string, 0, len(ids)),
	}
	for _, id := range ids {
		out.LiteratureIDs = append(out.LiteratureIDs, strconv.FormatUint(id, 10))
	}
	return out, nil
}

func (s *UserLiteratureService) SetCategoryLiteratureMembership(
	ctx context.Context,
	userID uint,
	categoryIDRaw string,
	literatureIDStrs []string,
) (*response.PaperUserLiteratureCategoryLiteratureIDsView, error) {
	categoryID, err := parseUserLiteratureCategoryID(categoryIDRaw)
	if err != nil {
		return nil, err
	}
	cat, err := s.categories.FindActiveByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, errorx.ErrParamsError.WithDetail("分类不存在")
	}
	litIDs, err := parseUserLiteratureIDList(literatureIDStrs)
	if err != nil {
		return nil, err
	}
	uid := uint64(userID)
	if err := s.repo.SyncCategoryLiteratureMembership(ctx, uid, categoryID, litIDs); err != nil {
		return nil, err
	}
	return s.ListCategoryLiteratureIDs(ctx, userID, categoryIDRaw)
}
