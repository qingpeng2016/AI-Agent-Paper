package paper

import (
	"context"
	"strconv"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
)

type LiteratureReviewService struct {
	reviews     repository.PaperOutputLiteratureReviewRepo
	manuscripts repository.PaperManuscriptRepo
}

func NewLiteratureReviewService(
	reviews repository.PaperOutputLiteratureReviewRepo,
	manuscripts repository.PaperManuscriptRepo,
) *LiteratureReviewService {
	return &LiteratureReviewService{reviews: reviews, manuscripts: manuscripts}
}

func (s *LiteratureReviewService) ListByManuscript(
	ctx context.Context,
	userID uint,
	manuscriptID uint64,
	page, pageSize int,
) (*response.PaperLiteratureReviewListView, error) {
	if manuscriptID == 0 {
		return nil, errorx.ErrParamsError
	}
	ms, err := s.manuscripts.GetByIDForUser(ctx, uint(manuscriptID), userID)
	if err != nil {
		return nil, err
	}
	if ms == nil {
		return nil, errorx.ErrParamsError.WithDetail("manuscript 不存在或无权访问")
	}
	page, pageSize, offset := NormalizeListPage(page, pageSize)
	total, err := s.reviews.CountByManuscriptForUser(ctx, userID, manuscriptID)
	if err != nil {
		return nil, err
	}
	rows, err := s.reviews.ListByManuscriptForUserPaged(ctx, userID, manuscriptID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	out := &response.PaperLiteratureReviewListView{
		ManuscriptID: manuscriptID,
		Items:        make([]response.PaperLiteratureReviewItemView, 0, len(rows)),
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
	}
	for _, row := range rows {
		out.Items = append(out.Items, literatureReviewItemView(row))
	}
	return out, nil
}

func (s *LiteratureReviewService) SoftDelete(
	ctx context.Context,
	userID uint,
	manuscriptID, reviewID uint64,
) error {
	if manuscriptID == 0 || reviewID == 0 {
		return errorx.ErrParamsError
	}
	ms, err := s.manuscripts.GetByIDForUser(ctx, uint(manuscriptID), userID)
	if err != nil {
		return err
	}
	if ms == nil {
		return errorx.ErrParamsError.WithDetail("manuscript 不存在或无权访问")
	}
	ok, err := s.reviews.UpdateStatusForManuscript(ctx, reviewID, manuscriptID, userID, "deleted")
	if err != nil {
		return err
	}
	if !ok {
		return errorx.ErrParamsError.WithDetail("文献综述不存在或无权操作")
	}
	return nil
}

func (s *LiteratureReviewService) GetDetail(
	ctx context.Context,
	userID uint,
	manuscriptID, reviewID uint64,
) (*response.PaperLiteratureReviewItemView, error) {
	if manuscriptID == 0 || reviewID == 0 {
		return nil, errorx.ErrParamsError
	}
	ms, err := s.manuscripts.GetByIDForUser(ctx, uint(manuscriptID), userID)
	if err != nil {
		return nil, err
	}
	if ms == nil {
		return nil, errorx.ErrParamsError.WithDetail("manuscript 不存在或无权访问")
	}
	row, err := s.reviews.GetByIDForUser(ctx, reviewID, manuscriptID, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errorx.ErrParamsError.WithDetail("文献综述不存在或无权访问")
	}
	if row.Status == "deleted" {
		return nil, errorx.ErrParamsError.WithDetail("文献综述已删除")
	}
	item := literatureReviewItemView(*row)
	return &item, nil
}

func literatureReviewItemView(row entity.PaperOutputLiteratureReview) response.PaperLiteratureReviewItemView {
	from := row.From
	if from == "" {
		from = "generated"
	}
	item := response.PaperLiteratureReviewItemView{
		ID:        strconv.FormatUint(row.ID, 10),
		Version:   row.Version,
		Status:    row.Status,
		From:      from,
		Format:    row.Format,
		CreatedAt: row.CreatedAt.UTC().Format(time.RFC3339),
	}
	if row.ExperimentPlanID != nil && *row.ExperimentPlanID > 0 {
		item.ExperimentPlanID = strconv.FormatUint(*row.ExperimentPlanID, 10)
	}
	if row.Title != nil {
		item.Title = *row.Title
	}
	if row.ContentMedium != nil {
		item.ContentMedium = *row.ContentMedium
	}
	if row.PdfURI != nil {
		item.PdfURI = *row.PdfURI
	}
	if row.WordURI != nil {
		item.WordURI = *row.WordURI
	}
	if row.PptURI != nil {
		item.PptURI = *row.PptURI
	}
	return item
}
