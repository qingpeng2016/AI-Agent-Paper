package handler

import (
	"net/http"
	"strconv"

	userAPIKeySvc "github.com/qingpeng2016/ai-token-mall/application/core-service/user_api_key"
	ginMiddleware "github.com/qingpeng2016/ai-token-mall/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type UserAPIKeyHandler struct {
	svc *userAPIKeySvc.Service
}

func NewUserAPIKeyHandler(svc *userAPIKeySvc.Service) *UserAPIKeyHandler {
	return &UserAPIKeyHandler{svc: svc}
}

func (h *UserAPIKeyHandler) ListMain(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.svc.ListMainKeys(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *UserAPIKeyHandler) ListTeam(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.svc.ListTeamSubKeys(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *UserAPIKeyHandler) CreateSub(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var req request.CreateSubAPIKeyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.CreateSubKey(c.Request.Context(), userID, &req)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *UserAPIKeyHandler) UpdateSubLimit(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || keyID == 0 {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var req request.UpdateSubAPIKeyLimitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	if err := h.svc.UpdateSubKeyLimit(c.Request.Context(), userID, uint(keyID), &req); err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, gin.H{"ok": true})
}

func (h *UserAPIKeyHandler) ListTeamMembers(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.svc.ListTeamMembers(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *UserAPIKeyHandler) ListAddableInvitees(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.svc.ListAddableInvitees(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *UserAPIKeyHandler) GetEnterpriseInquiry(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.svc.GetEnterpriseInquiryStatus(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *UserAPIKeyHandler) SubmitEnterpriseInquiry(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var req request.ApiTeamEnterpriseInquiryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.SubmitEnterpriseInquiry(c.Request.Context(), userID, &req)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *UserAPIKeyHandler) AddTeamMember(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var req request.AddApiTeamMemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	if err := h.svc.AddTeamMember(c.Request.Context(), userID, &req); err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, gin.H{"ok": true})
}
