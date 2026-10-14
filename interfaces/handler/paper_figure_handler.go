package handler

import (
	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type PaperFigureHandler struct {
	svc *papersvc.PaperFigureService
}

func NewPaperFigureHandler(svc *papersvc.PaperFigureService) *PaperFigureHandler {
	return &PaperFigureHandler{svc: svc}
}

// GetFigures GET /api/v1/paper/figures
func (h *PaperFigureHandler) GetFigures(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperFigureListQuery
	_ = c.ShouldBindQuery(&q)
	if q.Page < 1 {
		q.Page = 1
	}
	filter, err := papersvc.ParseFigureListFilter(q.Q, q.CategoryID, q.ManuscriptID)
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

// GetFigureCategories GET /api/v1/paper/figures/categories
func (h *PaperFigureHandler) GetFigureCategories(c *gin.Context) {
	_, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperFigureCategoryListQuery
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

// PostCreateFigureCategory POST /api/v1/paper/figures/categories
func (h *PaperFigureHandler) PostCreateFigureCategory(c *gin.Context) {
	_, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var req request.PaperFigureCategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.CreateCategory(c.Request.Context(), req.Name)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PutUpdateFigureCategory PUT /api/v1/paper/figures/categories
func (h *PaperFigureHandler) PutUpdateFigureCategory(c *gin.Context) {
	_, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var req request.PaperFigureCategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.UpdateCategory(c.Request.Context(), req.ID, req.Name)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// GetFigureCategoryFigureIDs GET /api/v1/paper/figures/categories/figure-ids
func (h *PaperFigureHandler) GetFigureCategoryFigureIDs(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperFigureCategoryFigureIDsQuery
	_ = c.ShouldBindQuery(&q)
	data, err := h.svc.ListCategoryFigureIDs(c.Request.Context(), userID, q.CategoryID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PutFigureCategoryFigure PUT /api/v1/paper/figures/categories/figure-membership
func (h *PaperFigureHandler) PutFigureCategoryFigure(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var req request.PaperFigureCategorySetFigureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.SetCategoryFigureMembership(
		c.Request.Context(),
		userID,
		req.CategoryID,
		req.FigureIDs,
	)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// GetManuscriptsWithFigures GET /api/v1/paper/figures/manuscripts-with-figures
func (h *PaperFigureHandler) GetManuscriptsWithFigures(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperManuscriptWithFiguresListQuery
	_ = c.ShouldBindQuery(&q)
	if q.Page < 1 {
		q.Page = 1
	}
	data, err := h.svc.ListManuscriptsWithFiguresPaged(c.Request.Context(), userID, q.Page, q.PageSize)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// GetManuscriptFigureIDs GET /api/v1/paper/figures/manuscripts/figure-ids
func (h *PaperFigureHandler) GetManuscriptFigureIDs(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperManuscriptFigureIDsQuery
	_ = c.ShouldBindQuery(&q)
	data, err := h.svc.ListManuscriptFigureIDs(c.Request.Context(), userID, q.ManuscriptID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PutManuscriptFigureMembership PUT /api/v1/paper/figures/manuscripts/figure-membership
func (h *PaperFigureHandler) PutManuscriptFigureMembership(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var req request.PaperManuscriptSetFigureMembershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.SetManuscriptFigureMembership(
		c.Request.Context(),
		userID,
		req.ManuscriptID,
		req.FigureIDs,
	)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
