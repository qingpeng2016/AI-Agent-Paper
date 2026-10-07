package handler

import (
	"strconv"
	"strings"

	papersvc "github.com/qingpeng2016/ai-token-mall/application/core-service/paper"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/gin-gonic/gin"
)

type PaperLiteratureHandler struct {
	lit *papersvc.LiteratureSearchService
}

func NewPaperLiteratureHandler(lit *papersvc.LiteratureSearchService) *PaperLiteratureHandler {
	return &PaperLiteratureHandler{lit: lit}
}

func (h *PaperLiteratureHandler) literatureQueryLimit(c *gin.Context) (string, int, bool) {
	query := strings.TrimSpace(c.Query("query"))
	if query == "" {
		response.ResponseErr(c, errorx.ErrParamsError)
		return "", 0, false
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	return query, limit, true
}

// SearchArxiv GET /api/v1/third-party/literature/arxiv?query=&limit=20
func (h *PaperLiteratureHandler) SearchArxiv(c *gin.Context) {
	query, limit, ok := h.literatureQueryLimit(c)
	if !ok {
		return
	}
	data, err := h.lit.SearchArxiv(c.Request.Context(), query, limit)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// SearchOpenAlex GET /api/v1/third-party/literature/openalex?query=&limit=20
func (h *PaperLiteratureHandler) SearchOpenAlex(c *gin.Context) {
	query, limit, ok := h.literatureQueryLimit(c)
	if !ok {
		return
	}
	data, err := h.lit.SearchOpenAlex(c.Request.Context(), query, limit)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// SearchSemanticScholar GET /api/v1/third-party/literature/semantic-scholar?query=&limit=20
func (h *PaperLiteratureHandler) SearchSemanticScholar(c *gin.Context) {
	query, limit, ok := h.literatureQueryLimit(c)
	if !ok {
		return
	}
	data, err := h.lit.SearchSemanticScholar(c.Request.Context(), query, limit)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
