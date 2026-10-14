package figure

import (
	"context"
	"strconv"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/listpage"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/userliterature"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
)

func (s *PaperFigureService) ListManuscriptsWithFiguresPaged(
	ctx context.Context,
	userID uint,
	page, pageSize int,
) (*response.PaperManuscriptWithFiguresListView, error) {
	page, pageSize, offset := listpage.NormalizeListPage(page, pageSize)
	total, err := s.manuscripts.CountByUserActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.manuscripts.ListByUserPaged(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	msIDs := make([]uint64, 0, len(rows))
	for _, row := range rows {
		if row.ID > 0 {
			msIDs = append(msIDs, uint64(row.ID))
		}
	}
	figIDByMs := map[uint64][]string{}
	figTitleByMs := map[uint64][]string{}
	if len(msIDs) > 0 {
		bindings, err := s.msFigureRel.ListActiveFigureBindingsByManuscriptIDs(ctx, userID, msIDs)
		if err != nil {
			return nil, err
		}
		for _, b := range bindings {
			figIDByMs[b.ManuscriptID] = append(figIDByMs[b.ManuscriptID], strconv.FormatUint(b.FigureID, 10))
			title := b.FigureTitle
			if title != "" {
				figTitleByMs[b.ManuscriptID] = append(figTitleByMs[b.ManuscriptID], title)
			}
		}
	}
	out := &response.PaperManuscriptWithFiguresListView{
		Items:    make([]response.PaperManuscriptWithFiguresItemView, 0, len(rows)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, row := range rows {
		mid := uint64(row.ID)
		item := response.PaperManuscriptWithFiguresItemView{
			ID:        strconv.FormatUint(mid, 10),
			Title:     row.Title,
			IsCurrent: row.IsCurrent,
		}
		item.FigureIDs = figIDByMs[mid]
		item.FigureTitles = figTitleByMs[mid]
		if !row.CreatedAt.IsZero() {
			item.CreatedAt = row.CreatedAt.UTC().Format(time.RFC3339)
		}
		out.Items = append(out.Items, item)
	}
	return out, nil
}

func (s *PaperFigureService) ListManuscriptFigureIDs(
	ctx context.Context,
	userID uint,
	manuscriptIDRaw string,
) (*response.PaperManuscriptFigureIDsView, error) {
	manuscriptID, err := userliterature.ParseManuscriptID(manuscriptIDRaw)
	if err != nil {
		return nil, err
	}
	ms, err := s.manuscripts.GetByIDForUser(ctx, uint(manuscriptID), userID)
	if err != nil {
		return nil, err
	}
	if ms == nil {
		return nil, errorx.ErrParamsError.WithDetail("论文不存在")
	}
	ids, err := s.msFigureRel.ListActiveFigureIDsByManuscript(ctx, userID, manuscriptID)
	if err != nil {
		return nil, err
	}
	out := &response.PaperManuscriptFigureIDsView{
		ManuscriptID: strconv.FormatUint(manuscriptID, 10),
		FigureIDs:    make([]string, 0, len(ids)),
	}
	for _, id := range ids {
		out.FigureIDs = append(out.FigureIDs, strconv.FormatUint(id, 10))
	}
	return out, nil
}

func (s *PaperFigureService) SetManuscriptFigureMembership(
	ctx context.Context,
	userID uint,
	manuscriptIDRaw string,
	figureIDStrs []string,
) (*response.PaperManuscriptFigureIDsView, error) {
	manuscriptID, err := userliterature.ParseManuscriptID(manuscriptIDRaw)
	if err != nil {
		return nil, err
	}
	ms, err := s.manuscripts.GetByIDForUser(ctx, uint(manuscriptID), userID)
	if err != nil {
		return nil, err
	}
	if ms == nil {
		return nil, errorx.ErrParamsError.WithDetail("论文不存在")
	}
	figIDs, err := parseFigureIDList(figureIDStrs)
	if err != nil {
		return nil, err
	}
	if err := s.msFigureRel.SyncManuscriptFigureMembership(ctx, userID, manuscriptID, figIDs); err != nil {
		return nil, err
	}
	return s.ListManuscriptFigureIDs(ctx, userID, manuscriptIDRaw)
}
