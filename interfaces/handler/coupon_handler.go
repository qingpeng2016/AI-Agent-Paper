package handler

import (
	"net/http"

	couponSvc "github.com/qingpeng2016/ai-token-mall/application/core-service/coupon"
	ginMiddleware "github.com/qingpeng2016/ai-token-mall/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/gin-gonic/gin"
)

type CouponHandler struct {
	svc *couponSvc.Service
}

func NewCouponHandler(svc *couponSvc.Service) *CouponHandler {
	return &CouponHandler{svc: svc}
}

func (h *CouponHandler) RegisterPromo(c *gin.Context) {
	data, err := h.svc.RegisterPromo(c.Request.Context())
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *CouponHandler) ListMine(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.svc.ListMine(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
