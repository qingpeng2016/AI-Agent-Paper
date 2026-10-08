package handler

import (
	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type PaperTopicDiscoveryHandler struct {
	opts *papersvc.TopicDiscoveryOptionsService
	run  *papersvc.TopicDiscoveryRunService
}

func NewPaperTopicDiscoveryHandler(
	opts *papersvc.TopicDiscoveryOptionsService,
	run *papersvc.TopicDiscoveryRunService,
) *PaperTopicDiscoveryHandler {
	return &PaperTopicDiscoveryHandler{opts: opts, run: run}
}

// GetFormOptions GET /api/v1/paper/topic-discovery/form-options
func (h *PaperTopicDiscoveryHandler) GetFormOptions(c *gin.Context) {
	data, err := h.opts.GetFormOptions(c.Request.Context())
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PostRun POST /api/v1/paper/topic-discovery/run
func (h *PaperTopicDiscoveryHandler) PostRun(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var req request.TopicDiscoveryRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.run.Run(c.Request.Context(), userID, req)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// GetCurrentRun GET /api/v1/paper/topic-discovery/run/current?manuscript_id=
func (h *PaperTopicDiscoveryHandler) GetCurrentRun(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.TopicDiscoveryRunQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.run.GetCurrentRun(c.Request.Context(), userID, q.ManuscriptID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
