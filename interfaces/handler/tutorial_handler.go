package handler

import (
	"errors"
	"net/http"

	"github.com/qingpeng2016/ai-token-mall/application/core-service/tutorial"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/gin-gonic/gin"
)

type TutorialHandler struct {
	tutorialSvc *tutorial.TutorialService
}

func NewTutorialHandler(tutorialSvc *tutorial.TutorialService) *TutorialHandler {
	return &TutorialHandler{tutorialSvc: tutorialSvc}
}

// List 教程列表（分类 + 文章摘要）
func (h *TutorialHandler) List(c *gin.Context) {
	categoryCode := c.Query("category_code")
	data, err := h.tutorialSvc.List(c.Request.Context(), categoryCode)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// Detail 教程正文
func (h *TutorialHandler) Detail(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "slug required"})
		return
	}
	data, err := h.tutorialSvc.Detail(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, tutorial.ErrTutorialNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "tutorial not found"})
			return
		}
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
