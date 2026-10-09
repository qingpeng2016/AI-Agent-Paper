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
	rows, err := s.reviews.ListByManuscript(ctx, manuscriptID)
	if err != nil {
		return nil, err
	}
	out := &response.PaperLiteratureReviewListView{
		ManuscriptID: manuscriptID,
		Items:        make([]response.PaperLiteratureReviewItemView, 0, len(rows)),
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

func literatureReviewItemView(row entity.PaperOutputLiteratureReview) response.PaperLiteratureReviewItemView {
	item := response.PaperLiteratureReviewItemView{
		ID:        strconv.FormatUint(row.ID, 10),
		Version:   row.Version,
		Status:    row.Status,
		Format:    row.Format,
		CreatedAt: row.CreatedAt.UTC().Format(time.RFC3339),
	}
	if row.Structure != nil {
		item.Structure = *row.Structure
	}
	if row.Title != nil {
		item.Title = *row.Title
	}
	if row.Summary != nil {
		item.Summary = *row.Summary
	}
	if row.ContentMedium != nil {
		item.ContentMedium = *row.ContentMedium
	}
	return item
}
