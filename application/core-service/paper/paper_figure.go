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

type PaperFigureService struct {
	repo          repository.PaperFigureRepo
	categories    repository.PaperFigureCategoryRepo
	manuscripts   repository.PaperManuscriptRepo
	msFigureRel   repository.PaperFigureManuscriptRelRepo
}

func NewPaperFigureService(
	repo repository.PaperFigureRepo,
	categories repository.PaperFigureCategoryRepo,
	manuscripts repository.PaperManuscriptRepo,
	msFigureRel repository.PaperFigureManuscriptRelRepo,
) *PaperFigureService {
	return &PaperFigureService{
		repo:        repo,
		categories:  categories,
		manuscripts: manuscripts,
		msFigureRel: msFigureRel,
	}
}

func (s *PaperFigureService) ListMinePaged(
	ctx context.Context,
	userID uint,
	page, pageSize int,
	filter repository.FigureListFilter,
) (*response.PaperFigureListView, error) {
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
	figIDs := make([]uint64, 0, len(rows))
	for _, row := range rows {
		if row.ID > 0 {
			figIDs = append(figIDs, row.ID)
		}
	}
	catByFig := map[uint64][]repository.FigureCategoryBinding{}
	msByFig := map[uint64][]repository.FigureManuscriptBinding{}
	if len(figIDs) > 0 {
		bindings, err := s.repo.ListActiveCategoryBindingsByFigureIDs(ctx, uid, figIDs)
		if err != nil {
			return nil, err
		}
		for _, b := range bindings {
			catByFig[b.FigureID] = append(catByFig[b.FigureID], b)
		}
		msBindings, err := s.msFigureRel.ListActiveManuscriptBindingsByFigureIDs(ctx, userID, figIDs)
		if err != nil {
			return nil, err
		}
		for _, b := range msBindings {
			msByFig[b.FigureID] = append(msByFig[b.FigureID], b)
		}
	}
	out := &response.PaperFigureListView{
		Items:    make([]response.PaperFigureItemView, 0, len(rows)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, row := range rows {
		fig := row.PaperFigure
		item := response.PaperFigureItemView{
			ID:       strconv.FormatUint(fig.ID, 10),
			Origin:   fig.Origin,
			FileKind: fig.FileKind,
			FileName: fig.FileName,
			Title:    fig.Title,
			SizeBytes: fig.SizeBytes,
		}
		for _, b := range catByFig[fig.ID] {
			item.CategoryIDs = append(item.CategoryIDs, strconv.FormatUint(b.CategoryID, 10))
			name := strings.TrimSpace(b.CategoryName)
			if name != "" {
				item.CategoryNames = append(item.CategoryNames, name)
			}
		}
		for _, b := range msByFig[fig.ID] {
			item.ManuscriptIDs = append(item.ManuscriptIDs, strconv.FormatUint(b.ManuscriptID, 10))
			title := strings.TrimSpace(b.ManuscriptTitle)
			if title != "" {
				item.ManuscriptTitles = append(item.ManuscriptTitles, title)
			}
		}
		if fig.Note != nil {
			item.Note = *fig.Note
		}
		if fig.StorageURI != nil {
			item.StorageURI = *fig.StorageURI
		}
		if !fig.CreatedAt.IsZero() {
			item.CreatedAt = fig.CreatedAt.UTC().Format(time.RFC3339)
		}
		out.Items = append(out.Items, item)
	}
	return out, nil
}

func (s *PaperFigureService) ListCategoriesPaged(
	ctx context.Context,
	page, pageSize int,
) (*response.PaperFigureCategoryListView, error) {
	page, pageSize, offset := NormalizeListPage(page, pageSize)
	total, err := s.categories.CountActive(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.categories.ListActivePaged(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}
	out := &response.PaperFigureCategoryListView{
		Items:    make([]response.PaperFigureCategoryItemView, 0, len(rows)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, row := range rows {
		item := response.PaperFigureCategoryItemView{
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

func (s *PaperFigureService) CreateCategory(
	ctx context.Context,
	name string,
) (*response.PaperFigureCategoryItemView, error) {
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
	row := &entity.PaperFigureCategory{
		Name:   name,
		Status: "active",
	}
	if err := s.categories.Create(ctx, row); err != nil {
		return nil, err
	}
	return figureCategoryItemView(row), nil
}

func (s *PaperFigureService) UpdateCategory(
	ctx context.Context,
	categoryIDRaw string,
	name string,
) (*response.PaperFigureCategoryItemView, error) {
	categoryID, err := parseFigureCategoryID(categoryIDRaw)
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
		return figureCategoryItemView(cat), nil
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
	return figureCategoryItemView(updated), nil
}

func figureCategoryItemView(row *entity.PaperFigureCategory) *response.PaperFigureCategoryItemView {
	if row == nil {
		return nil
	}
	item := &response.PaperFigureCategoryItemView{
		ID:     strconv.FormatUint(row.ID, 10),
		Name:   row.Name,
		Status: row.Status,
	}
	if !row.CreatedAt.IsZero() {
		item.CreatedAt = row.CreatedAt.UTC().Format(time.RFC3339)
	}
	return item
}

func (s *PaperFigureService) ListCategoryFigureIDs(
	ctx context.Context,
	userID uint,
	categoryIDRaw string,
) (*response.PaperFigureCategoryFigureIDsView, error) {
	categoryID, err := parseFigureCategoryID(categoryIDRaw)
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
	ids, err := s.repo.ListActiveFigureIDsByCategory(ctx, uid, categoryID)
	if err != nil {
		return nil, err
	}
	out := &response.PaperFigureCategoryFigureIDsView{
		CategoryID: strconv.FormatUint(categoryID, 10),
		FigureIDs:  make([]string, 0, len(ids)),
	}
	for _, id := range ids {
		out.FigureIDs = append(out.FigureIDs, strconv.FormatUint(id, 10))
	}
	return out, nil
}

func (s *PaperFigureService) SetCategoryFigureMembership(
	ctx context.Context,
	userID uint,
	categoryIDRaw string,
	figureIDStrs []string,
) (*response.PaperFigureCategoryFigureIDsView, error) {
	categoryID, err := parseFigureCategoryID(categoryIDRaw)
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
	figIDs, err := parseFigureIDList(figureIDStrs)
	if err != nil {
		return nil, err
	}
	uid := uint64(userID)
	if err := s.repo.SyncCategoryFigureMembership(ctx, uid, categoryID, figIDs); err != nil {
		return nil, err
	}
	return s.ListCategoryFigureIDs(ctx, userID, categoryIDRaw)
}
