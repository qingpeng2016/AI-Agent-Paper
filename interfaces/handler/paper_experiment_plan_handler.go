package handler

import (
	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type PaperExperimentPlanHandler struct {
	svc *papersvc.ExperimentPlanService
}

func NewPaperExperimentPlanHandler(svc *papersvc.ExperimentPlanService) *PaperExperimentPlanHandler {
	return &PaperExperimentPlanHandler{svc: svc}
}

// GetExperimentPlans GET /api/v1/paper/experiment-plans?manuscript_id=
func (h *PaperExperimentPlanHandler) GetExperimentPlans(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperExperimentPlanListQuery
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

// GetExperimentPlanDetail GET /api/v1/paper/experiment-plans/detail?manuscript_id=&experiment_plan_id=
func (h *PaperExperimentPlanHandler) GetExperimentPlanDetail(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperExperimentPlanDetailQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.GetDetail(
		c.Request.Context(),
		userID,
		q.ManuscriptID,
		q.ExperimentPlanID,
	)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PostUploadExperimentData POST /api/v1/paper/experiment-plans/upload-experiment-data
func (h *PaperExperimentPlanHandler) PostUploadExperimentData(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperExperimentPlanUploadForm
	if err := c.ShouldBind(&q); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseErr(c, errorx.ErrParamsError.WithDetail("请选择文件"))
		return
	}
	f, err := file.Open()
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	defer f.Close()
	data, err := h.svc.UploadExperimentData(
		c.Request.Context(),
		userID,
		q.ManuscriptID,
		q.ExperimentPlanID,
		file.Filename,
		f,
	)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PostDeleteExperimentData POST /api/v1/paper/experiment-plans/delete-experiment-data
func (h *PaperExperimentPlanHandler) PostDeleteExperimentData(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var body request.PaperExperimentPlanDeleteExperimentDataBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.DeleteExperimentData(
		c.Request.Context(),
		userID,
		body.ManuscriptID,
		body.ExperimentPlanID,
	)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
