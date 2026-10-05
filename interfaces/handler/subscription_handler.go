package handler

import (
	"net/http"

	"github.com/qingpeng2016/ai-token-mall/application/core-service/subscription"
	ginMiddleware "github.com/qingpeng2016/ai-token-mall/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	subSvc *subscription.SubscriptionService
}

func NewSubscriptionHandler(subSvc *subscription.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subSvc: subSvc}
}

func (h *SubscriptionHandler) ListMine(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.subSvc.ListMine(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
