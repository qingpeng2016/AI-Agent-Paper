package handler

import (
	"io"
	"net/http"
	"strings"

	inviteRebateSvc "github.com/qingpeng2016/ai-token-mall/application/core-service/invite_rebate"
	ginMiddleware "github.com/qingpeng2016/ai-token-mall/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type InviteRebateHandler struct {
	svc *inviteRebateSvc.Service
}

func NewInviteRebateHandler(svc *inviteRebateSvc.Service) *InviteRebateHandler {
	return &InviteRebateHandler{svc: svc}
}

func (h *InviteRebateHandler) Overview(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.svc.GetOverview(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *InviteRebateHandler) ListMembers(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.svc.ListMembers(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *InviteRebateHandler) ListCommissionRecords(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var q request.InviteRebatePageQuery
	_ = c.ShouldBindQuery(&q)
	data, err := h.svc.ListCommissionRecords(c.Request.Context(), userID, &q)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *InviteRebateHandler) ListWithdrawals(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var q request.InviteRebatePageQuery
	_ = c.ShouldBindQuery(&q)
	data, err := h.svc.ListWithdrawals(c.Request.Context(), userID, &q)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *InviteRebateHandler) GetPayoutConfig(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.svc.GetPayoutConfig(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *InviteRebateHandler) UploadPayoutQR(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	channel := strings.TrimSpace(c.PostForm("channel"))
	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	f, err := file.Open()
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	defer f.Close()
	const maxQR = 2 << 20
	data, err := io.ReadAll(io.LimitReader(f, maxQR+1))
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	mime := strings.TrimSpace(file.Header.Get("Content-Type"))
	if i := strings.Index(mime, ";"); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}
	if mime == "" {
		mime = http.DetectContentType(data)
		if i := strings.Index(mime, ";"); i >= 0 {
			mime = mime[:i]
		}
	}
	if err := h.svc.SavePayoutQR(c.Request.Context(), userID, channel, data, mime); err != nil {
		response.ResponseErr(c, err)
		return
	}
	cfg, err := h.svc.GetPayoutConfig(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, cfg)
}

func (h *InviteRebateHandler) CreateWithdrawal(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var req request.CreateCommissionWithdrawalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	data, err := h.svc.CreateCommissionWithdrawal(c.Request.Context(), userID, req.Amount, req.Channel)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *InviteRebateHandler) TransferCommissionToBalance(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var req request.CommissionTransferToBalanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	data, err := h.svc.TransferCommissionToBalance(c.Request.Context(), userID, req.Amount)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
