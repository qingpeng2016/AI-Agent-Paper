package handler

import (
	"net/http"

	"github.com/qingpeng2016/ai-token-mall/application/core-service/user"
	"github.com/qingpeng2016/ai-token-mall/common/constants"
	ginMiddleware "github.com/qingpeng2016/ai-token-mall/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userSvc *user.UserService
}

func NewUserHandler(userSvc *user.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req request.RegisterUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	data, err := h.userSvc.Register(c.Request.Context(), &req)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		constants.AuthTokenCookie,
		data.Token,
		constants.AuthTokenMaxAge,
		"/",
		"",
		false,
		false,
	)
	payload := gin.H{"user": data.User}
	if data.RegisterCouponsGranted > 0 {
		payload["register_coupons_granted"] = data.RegisterCouponsGranted
	}
	response.ResponseSuccess(c, payload)
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req request.LoginUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	data, err := h.userSvc.Login(c.Request.Context(), &req)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		constants.AuthTokenCookie,
		data.Token,
		constants.AuthTokenMaxAge,
		"/",
		"",
		false,
		false,
	)
	response.ResponseSuccess(c, gin.H{"user": data.User})
}

// Me 当前登录用户信息（会员中心概览 / 账户设置）
func (h *UserHandler) Me(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.userSvc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// ChangePassword 修改登录密码（需验证原密码）
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var req request.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	if err := h.userSvc.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, nil)
}

// ListWalletFlows 当前用户资金流水（会员中心 · 资金流水）
func (h *UserHandler) ListWalletFlows(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var q request.ListWalletFlowsQuery
	_ = c.ShouldBindQuery(&q)
	data, err := h.userSvc.ListWalletFlows(c.Request.Context(), userID, &q)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// ListInvoices 当前用户发票列表（会员中心 · 发票管理）
func (h *UserHandler) ListInvoices(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var q request.ListInvoicesQuery
	_ = c.ShouldBindQuery(&q)
	data, err := h.userSvc.ListInvoices(c.Request.Context(), userID, &q)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// Logout 退出登录：清空 atm_token Cookie
func (h *UserHandler) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(constants.AuthTokenCookie, "", -1, "/", "", false, false)
	response.ResponseSuccess(c, nil)
}
