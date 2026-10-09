package handler

import (
	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type PaperLiteratureReviewHandler struct {
	svc      *papersvc.LiteratureReviewService
	expPlans *papersvc.ExperimentPlanService
}

func NewPaperLiteratureReviewHandler(
	svc *papersvc.LiteratureReviewService,
	expPlans *papersvc.ExperimentPlanService,
) *PaperLiteratureReviewHandler {
	return &PaperLiteratureReviewHandler{svc: svc, expPlans: expPlans}
}

// GetLiteratureReviews GET /api/v1/paper/literature-reviews?manuscript_id=
func (h *PaperLiteratureReviewHandler) GetLiteratureReviews(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperLiteratureReviewListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.ListByManuscript(c.Request.Context(), userID, q.ManuscriptID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PostSoftDeleteLiteratureReview POST /api/v1/paper/literature-reviews/soft-delete
func (h *PaperLiteratureReviewHandler) PostSoftDeleteLiteratureReview(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var body request.PaperLiteratureReviewSoftDeleteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	if err := h.svc.SoftDelete(
		c.Request.Context(),
		userID,
		body.ManuscriptID,
		body.LiteratureReviewID,
	); err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, gin.H{"ok": true})
}

// PostGenerateExperimentPlan POST /api/v1/paper/literature-reviews/generate-experiment-plan
func (h *PaperLiteratureReviewHandler) PostGenerateExperimentPlan(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var body request.PaperLiteratureReviewGenerateExperimentPlanBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	if h.expPlans == nil {
		response.ResponseErr(c, errorx.ErrParamsError.WithDetail("实验方案服务未配置"))
		return
	}
	if err := h.expPlans.EnqueueGenerate(
		c.Request.Context(),
		userID,
		body.ManuscriptID,
		body.LiteratureReviewID,
	); err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, gin.H{
		"ok":     true,
		"status": papersvc.LitReviewStatusGeneratingExperimentPlan,
	})
}

// GetLiteratureReviewDetail GET /api/v1/paper/literature-reviews/detail?manuscript_id=&literature_review_id=
func (h *PaperLiteratureReviewHandler) GetLiteratureReviewDetail(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperLiteratureReviewDetailQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.GetDetail(
		c.Request.Context(),
		userID,
		q.ManuscriptID,
		q.LiteratureReviewID,
	)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
