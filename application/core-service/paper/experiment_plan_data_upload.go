package paper

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
)

const maxExperimentDataUploadBytes = 50 << 20 // 50MB

func (s *ExperimentPlanService) UploadExperimentData(
	ctx context.Context,
	userID uint,
	manuscriptID, planID uint64,
	originalFileName string,
	content io.Reader,
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

	rel := ExperimentDataLocalRelPath(manuscriptID, planID, originalFileName)
	rel = filepath.ToSlash(rel)
	abs := StorageLocalAbsPath(rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return nil, err
	}
	tmp := abs + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return nil, err
	}
	n, copyErr := io.Copy(f, io.LimitReader(content, maxExperimentDataUploadBytes+1))
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return nil, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return nil, closeErr
	}
	if n > maxExperimentDataUploadBytes {
		_ = os.Remove(tmp)
		return nil, errorx.ErrParamsError.WithDetail(fmt.Sprintf("文件超过 %dMB 限制", maxExperimentDataUploadBytes>>20))
	}
	if n == 0 {
		_ = os.Remove(tmp)
		return nil, errorx.ErrParamsError.WithDetail("文件为空")
	}
	if err := os.Rename(tmp, abs); err != nil {
		_ = os.Remove(tmp)
		return nil, err
	}

	ok, err := s.plans.UpdateExperimentDataURI(ctx, planID, manuscriptID, userID, rel)
	if err != nil {
		return nil, err
	}
	if !ok {
		_ = os.Remove(abs)
		return nil, errorx.ErrParamsError.WithDetail("更新实验数据路径失败")
	}

	// 替换上传时删除旧文件（若路径不同）
	if row.ExperimentDataURI != nil {
		old := strings.TrimSpace(*row.ExperimentDataURI)
		if old != "" && old != rel {
			_ = os.Remove(StorageLocalAbsPath(old))
		}
	}

	updated, err := s.plans.GetByIDForManuscript(ctx, planID, manuscriptID)
	if err != nil || updated == nil {
		item := experimentPlanItemView(*row)
		item.ExperimentDataURI = rel
		item = s.enrichExperimentPlanLiteratureReviewID(ctx, manuscriptID, planID, item)
		return &item, nil
	}
	item := experimentPlanItemView(*updated)
	item = s.enrichExperimentPlanLiteratureReviewID(ctx, manuscriptID, planID, item)
	return &item, nil
}
