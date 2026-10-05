package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/application/core-service/invoice"
	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	ginMiddleware "github.com/qingpeng2016/ai-token-mall/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type InvoiceConfigHandler struct {
	svc      *invoice.InvoiceConfigService
	lookup   *invoice.EnterpriseLookup
}

func NewInvoiceConfigHandler(svc *invoice.InvoiceConfigService, lookup *invoice.EnterpriseLookup) *InvoiceConfigHandler {
	return &InvoiceConfigHandler{svc: svc, lookup: lookup}
}

func (h *InvoiceConfigHandler) ListMine(c *gin.Context) {
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

func (h *InvoiceConfigHandler) Create(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var req request.SaveInvoiceConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	data, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *InvoiceConfigHandler) Update(c *gin.Context) {
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
	var req request.SaveInvoiceConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	data, err := h.svc.Update(c.Request.Context(), userID, uint(id), &req)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *InvoiceConfigHandler) SetDefault(c *gin.Context) {
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
	if err := h.svc.SetDefault(c.Request.Context(), userID, uint(id)); err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, nil)
}

func (h *InvoiceConfigHandler) LookupEnterprise(c *gin.Context) {
	_, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.lookup.Lookup(c.Request.Context(), keyword)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
