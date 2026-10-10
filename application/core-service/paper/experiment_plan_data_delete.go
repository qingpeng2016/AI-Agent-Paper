package paper

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
)

func (s *ExperimentPlanService) DeleteExperimentData(
	ctx context.Context,
	userID uint,
	manuscriptID, planID uint64,
) (*response.PaperExperimentPlanItemView, error) {
	if manuscriptID == 0 || planID == 0 {
		return nil, errorx.ErrParamsError
	}
	ms, err := s.manuscripts.GetByIDForUser(ctx, uint(manuscriptID), userID)
	if err != nil {
		return nil, err
	}
	if ms == nil {
		return nil, errorx.ErrParamsError.WithDetail("manuscript 不存在或无权访问")
	}
	row, err := s.plans.GetByIDForUser(ctx, planID, manuscriptID, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errorx.ErrParamsError.WithDetail("实验方案不存在或无权访问")
	}
	if row.Status == "deleted" {
		return nil, errorx.ErrParamsError.WithDetail("实验方案已删除")
	}

	rel := ""
	if row.ExperimentDataURI != nil {
		rel = strings.TrimSpace(*row.ExperimentDataURI)
	}
	if rel == "" {
		return nil, errorx.ErrParamsError.WithDetail("暂无实验数据")
	}

	ok, err := s.plans.ClearExperimentDataURI(ctx, planID, manuscriptID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errorx.ErrParamsError.WithDetail("清除实验数据记录失败")
	}

	relSlash := filepath.ToSlash(rel)
	if strings.HasPrefix(relSlash, experimentDataStorageSubdir+"/") {
		_ = os.Remove(StorageLocalAbsPath(relSlash))
	}

	updated, err := s.plans.GetByIDForManuscript(ctx, planID, manuscriptID)
	if err != nil || updated == nil {
		item := experimentPlanItemView(*row)
		item.ExperimentDataURI = ""
		item = s.enrichExperimentPlanLiteratureReviewID(ctx, manuscriptID, planID, item)
		return &item, nil
	}
	item := experimentPlanItemView(*updated)
	item = s.enrichExperimentPlanLiteratureReviewID(ctx, manuscriptID, planID, item)
	return &item, nil
}
