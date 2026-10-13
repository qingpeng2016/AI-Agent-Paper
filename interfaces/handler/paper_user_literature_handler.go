package handler

import (
	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type PaperUserLiteratureHandler struct {
	svc *papersvc.UserLiteratureService
}

func NewPaperUserLiteratureHandler(svc *papersvc.UserLiteratureService) *PaperUserLiteratureHandler {
	return &PaperUserLiteratureHandler{svc: svc}
}

// GetUserLiterature GET /api/v1/paper/user-literature
func (h *PaperUserLiteratureHandler) GetUserLiterature(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperUserLiteratureListQuery
	_ = c.ShouldBindQuery(&q)
	if q.Page < 1 {
		q.Page = 1
	}
	filter, err := papersvc.ParseUserLiteratureListFilter(q.Q, q.CategoryID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	data, err := h.svc.ListMinePaged(c.Request.Context(), userID, q.Page, q.PageSize, filter)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// GetUserLiteratureCategories GET /api/v1/paper/user-literature/categories
func (h *PaperUserLiteratureHandler) GetUserLiteratureCategories(c *gin.Context) {
	_, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperUserLiteratureCategoryListQuery
	_ = c.ShouldBindQuery(&q)
	if q.Page < 1 {
		q.Page = 1
	}
	data, err := h.svc.ListCategoriesPaged(c.Request.Context(), q.Page, q.PageSize)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
