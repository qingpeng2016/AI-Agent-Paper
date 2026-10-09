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
	svc *papersvc.LiteratureReviewService
}

func NewPaperLiteratureReviewHandler(svc *papersvc.LiteratureReviewService) *PaperLiteratureReviewHandler {
	return &PaperLiteratureReviewHandler{svc: svc}
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
