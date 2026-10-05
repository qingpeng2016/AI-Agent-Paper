package handler

import (
	"errors"
	"net/http"

	"github.com/qingpeng2016/ai-token-mall/application/core-service/enterprise"
	ginMiddleware "github.com/qingpeng2016/ai-token-mall/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type EnterpriseHandler struct {
	enterpriseSvc *enterprise.EnterpriseService
}

func NewEnterpriseHandler(enterpriseSvc *enterprise.EnterpriseService) *EnterpriseHandler {
	return &EnterpriseHandler{enterpriseSvc: enterpriseSvc}
}

// SubmitInquiry 企业采购需求提交
func (h *EnterpriseHandler) SubmitInquiry(c *gin.Context) {
	var req request.SubmitEnterpriseInquiryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	var userID *uint
	if uid, ok := ginMiddleware.TryUserIDFromRequest(c); ok {
		userID = &uid
	}
	id, err := h.enterpriseSvc.SubmitInquiry(c.Request.Context(), &req, userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, gin.H{
		"id":      id,
		"message": "已收到采购需求，顾问将尽快联系您",
	})
}

// GetProduct 企业方案详情
func (h *EnterpriseHandler) GetProduct(c *gin.Context) {
	code := c.Param("code")
	data, err := h.enterpriseSvc.GetProduct(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, enterprise.ErrEnterpriseProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "enterprise product not found"})
			return
		}
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// ListProducts 企业方案列表
func (h *EnterpriseHandler) ListProducts(c *gin.Context) {
	data, err := h.enterpriseSvc.ListProducts(c.Request.Context())
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
