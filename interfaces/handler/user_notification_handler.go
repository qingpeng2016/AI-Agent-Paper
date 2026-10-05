package handler

import (
	"net/http"
	"strconv"

	notifsvc "github.com/qingpeng2016/ai-token-mall/application/core-service/notification"
	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	ginMiddleware "github.com/qingpeng2016/ai-token-mall/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type UserNotificationHandler struct {
	svc *notifsvc.UserNotificationService
}

func NewUserNotificationHandler(svc *notifsvc.UserNotificationService) *UserNotificationHandler {
	return &UserNotificationHandler{svc: svc}
}

func (h *UserNotificationHandler) ListMine(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var q request.ListUserNotificationsQuery
	_ = c.ShouldBindQuery(&q)
	data, err := h.svc.ListMine(c.Request.Context(), userID, &q)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *UserNotificationHandler) UnreadCount(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.svc.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *UserNotificationHandler) MarkRead(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	if err := h.svc.MarkRead(c.Request.Context(), userID, uint(id)); err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, nil)
}

func (h *UserNotificationHandler) MarkAllRead(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	if err := h.svc.MarkAllRead(c.Request.Context(), userID); err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, nil)
}
