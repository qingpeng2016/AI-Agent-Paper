package userliterature

import (
	"context"
	"strconv"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/listpage"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
)

func (s *Service) ListManuscriptsWithLiteraturePaged(
	ctx context.Context,
	userID uint,
	page, pageSize int,
) (*response.PaperManuscriptWithLiteratureListView, error) {
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
	litIDByMs := map[uint64][]string{}
	litTitleByMs := map[uint64][]string{}
	if len(msIDs) > 0 {
		bindings, err := s.msLiteratureRel.ListActiveLiteratureBindingsByManuscriptIDs(ctx, userID, msIDs)
		if err != nil {
			return nil, err
		}
		for _, b := range bindings {
			litIDByMs[b.ManuscriptID] = append(litIDByMs[b.ManuscriptID], strconv.FormatUint(b.UserLiteratureID, 10))
			title := b.LiteratureTitle
			if title != "" {
				litTitleByMs[b.ManuscriptID] = append(litTitleByMs[b.ManuscriptID], title)
			}
		}
	}
	out := &response.PaperManuscriptWithLiteratureListView{
		Items:    make([]response.PaperManuscriptWithLiteratureItemView, 0, len(rows)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, row := range rows {
		mid := uint64(row.ID)
		item := response.PaperManuscriptWithLiteratureItemView{
			ID:        strconv.FormatUint(mid, 10),
			Title:     row.Title,
			IsCurrent: row.IsCurrent,
		}
		item.LiteratureIDs = litIDByMs[mid]
		item.LiteratureTitles = litTitleByMs[mid]
		if !row.CreatedAt.IsZero() {
			item.CreatedAt = row.CreatedAt.UTC().Format(time.RFC3339)
		}
		out.Items = append(out.Items, item)
	}
	return out, nil
}

func (s *Service) ListManuscriptLiteratureIDs(
	ctx context.Context,
	userID uint,
	manuscriptIDRaw string,
) (*response.PaperManuscriptLiteratureIDsView, error) {
	manuscriptID, err := ParseManuscriptID(manuscriptIDRaw)
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
	ids, err := s.msLiteratureRel.ListActiveLiteratureIDsByManuscript(ctx, userID, manuscriptID)
	if err != nil {
		return nil, err
	}
	out := &response.PaperManuscriptLiteratureIDsView{
		ManuscriptID:  strconv.FormatUint(manuscriptID, 10),
		LiteratureIDs: make([]string, 0, len(ids)),
	}
	for _, id := range ids {
		out.LiteratureIDs = append(out.LiteratureIDs, strconv.FormatUint(id, 10))
	}
	return out, nil
}

func (s *Service) SetManuscriptLiteratureMembership(
	ctx context.Context,
	userID uint,
	manuscriptIDRaw string,
	literatureIDStrs []string,
) (*response.PaperManuscriptLiteratureIDsView, error) {
	manuscriptID, err := ParseManuscriptID(manuscriptIDRaw)
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
	litIDs, err := ParseUserLiteratureIDList(literatureIDStrs)
	if err != nil {
		return nil, err
	}
	if err := s.msLiteratureRel.SyncManuscriptLiteratureMembership(ctx, userID, manuscriptID, litIDs); err != nil {
		return nil, err
	}
	return s.ListManuscriptLiteratureIDs(ctx, userID, manuscriptIDRaw)
}
