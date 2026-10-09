package paper

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
	"gorm.io/gorm"
)

type ManuscriptService struct {
	manuscripts repository.PaperManuscriptRepo
	catalog     repository.PaperRefCatalogRepo
}

func NewManuscriptService(
	manuscripts repository.PaperManuscriptRepo,
	catalog repository.PaperRefCatalogRepo,
) *ManuscriptService {
	return &ManuscriptService{manuscripts: manuscripts, catalog: catalog}
}

func (s *ManuscriptService) ListMine(ctx context.Context, userID uint) (*response.PaperManuscriptListView, error) {
	rows, err := s.manuscripts.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := &response.PaperManuscriptListView{
		Items: make([]response.PaperManuscriptItemView, 0, len(rows)),
	}
	for _, row := range rows {
		item := s.manuscriptItemView(ctx, row)
		out.Items = append(out.Items, item)
		if row.IsCurrent {
			out.CurrentManuscriptID = uint64(row.ID)
		}
	}
	return out, nil
}

func (s *ManuscriptService) manuscriptItemView(ctx context.Context, row entity.PaperManuscript) response.PaperManuscriptItemView {
	item := response.PaperManuscriptItemView{
		ID:        strconv.FormatUint(uint64(row.ID), 10),
		Title:     row.Title,
		Status:    row.Status,
		IsCurrent: row.IsCurrent,
	}
	if row.DisciplineID != nil && *row.DisciplineID > 0 {
		disc, err := s.catalog.FindDisciplineByID(ctx, *row.DisciplineID)
		if err == nil && disc != nil {
			item.DisciplineCode = disc.Code
			item.DisciplineLabel = disc.Name
		}
	}
	return item
}

func (s *ManuscriptService) Create(
	ctx context.Context,
	userID uint,
	title, disciplineCode string,
) (*response.PaperManuscriptListView, error) {
	title = strings.TrimSpace(title)
	disciplineCode = strings.TrimSpace(disciplineCode)
	if title == "" || disciplineCode == "" {
		return nil, errorx.ErrParamsError
	}
	if len([]rune(title)) > 256 {
		title = string([]rune(title)[:256])
	}
	disc, err := s.catalog.FindDisciplineByCode(ctx, disciplineCode)
	if err != nil {
		return nil, err
	}
	if disc == nil {
		return nil, errorx.ErrParamsError.WithDetail("无效学科：" + disciplineCode)
	}
	discID := uint64(disc.ID)
	ms, err := s.manuscripts.Create(ctx, userID, title, &discID)
	if err != nil {
		return nil, err
	}
	if err := s.manuscripts.SetCurrentForUser(ctx, userID, ms.ID); err != nil {
		return nil, err
	}
	return s.ListMine(ctx, userID)
}

func (s *ManuscriptService) SetCurrent(ctx context.Context, userID uint, manuscriptID uint64) (*response.PaperManuscriptListView, error) {
	if manuscriptID == 0 {
		return nil, errorx.ErrParamsError
	}
	if err := s.manuscripts.SetCurrentForUser(ctx, userID, uint(manuscriptID)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.ErrParamsError.WithDetail(fmt.Sprintf("manuscript_id=%d 不存在或已归档", manuscriptID))
		}
		return nil, err
	}
	return s.ListMine(ctx, userID)
}
