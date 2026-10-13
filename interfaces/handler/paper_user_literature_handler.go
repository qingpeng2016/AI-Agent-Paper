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
	filter, err := papersvc.ParseUserLiteratureListFilter(q.Q, q.CategoryID, q.ManuscriptID)
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

// PostCreateUserLiteratureCategory POST /api/v1/paper/user-literature/categories
func (h *PaperUserLiteratureHandler) PostCreateUserLiteratureCategory(c *gin.Context) {
	_, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var req request.PaperUserLiteratureCategoryCreateRequest
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

// PutUpdateUserLiteratureCategory PUT /api/v1/paper/user-literature/categories
func (h *PaperUserLiteratureHandler) PutUpdateUserLiteratureCategory(c *gin.Context) {
	_, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var req request.PaperUserLiteratureCategoryUpdateRequest
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

// GetUserLiteratureCategoryLiteratureIDs GET /api/v1/paper/user-literature/categories/literature-ids
func (h *PaperUserLiteratureHandler) GetUserLiteratureCategoryLiteratureIDs(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperUserLiteratureCategoryLiteratureIDsQuery
	_ = c.ShouldBindQuery(&q)
	data, err := h.svc.ListCategoryLiteratureIDs(c.Request.Context(), userID, q.CategoryID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PutUserLiteratureCategoryLiterature PUT /api/v1/paper/user-literature/categories/literature-membership
func (h *PaperUserLiteratureHandler) PutUserLiteratureCategoryLiterature(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var req request.PaperUserLiteratureCategorySetLiteratureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.SetCategoryLiteratureMembership(
		c.Request.Context(),
		userID,
		req.CategoryID,
		req.LiteratureIDs,
	)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// GetManuscriptsWithLiterature GET /api/v1/paper/user-literature/manuscripts-with-literature
func (h *PaperUserLiteratureHandler) GetManuscriptsWithLiterature(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperManuscriptWithLiteratureListQuery
	_ = c.ShouldBindQuery(&q)
	if q.Page < 1 {
		q.Page = 1
	}
	data, err := h.svc.ListManuscriptsWithLiteraturePaged(c.Request.Context(), userID, q.Page, q.PageSize)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// GetManuscriptLiteratureIDs GET /api/v1/paper/user-literature/manuscripts/literature-ids
func (h *PaperUserLiteratureHandler) GetManuscriptLiteratureIDs(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperManuscriptLiteratureIDsQuery
	_ = c.ShouldBindQuery(&q)
	data, err := h.svc.ListManuscriptLiteratureIDs(c.Request.Context(), userID, q.ManuscriptID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PutManuscriptLiteratureMembership PUT /api/v1/paper/user-literature/manuscripts/literature-membership
func (h *PaperUserLiteratureHandler) PutManuscriptLiteratureMembership(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var req request.PaperManuscriptSetLiteratureMembershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.SetManuscriptLiteratureMembership(
		c.Request.Context(),
		userID,
		req.ManuscriptID,
		req.LiteratureIDs,
	)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
