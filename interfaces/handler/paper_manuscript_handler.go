package handler

import (
	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type PaperManuscriptHandler struct {
	ms *papersvc.ManuscriptService
}

func NewPaperManuscriptHandler(ms *papersvc.ManuscriptService) *PaperManuscriptHandler {
	return &PaperManuscriptHandler{ms: ms}
}

// GetManuscripts GET /api/v1/paper/manuscripts
func (h *PaperManuscriptHandler) GetManuscripts(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.ms.ListMine(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PostCreateManuscript POST /api/v1/paper/manuscripts
func (h *PaperManuscriptHandler) PostCreateManuscript(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var req request.PaperManuscriptCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.ms.Create(c.Request.Context(), userID, req.Title, req.DisciplineCode)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PostSetCurrentManuscript POST /api/v1/paper/manuscripts/set-current
func (h *PaperManuscriptHandler) PostSetCurrentManuscript(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var req request.PaperManuscriptSetCurrentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.ms.SetCurrent(c.Request.Context(), userID, req.ManuscriptID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
