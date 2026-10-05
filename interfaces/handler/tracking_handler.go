package handler

import (
	trackingSvc "github.com/qingpeng2016/ai-token-mall/application/core-service/tracking"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type TrackingHandler struct {
	svc *trackingSvc.Service
}

func NewTrackingHandler(svc *trackingSvc.Service) *TrackingHandler {
	return &TrackingHandler{svc: svc}
}

func (h *TrackingHandler) ReportEvents(c *gin.Context) {
	var req request.ReportTrackEventsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	data, err := h.svc.ReportEvents(c, &req)
	if err != nil {
		response.ResponseErr(c, errorx.ErrDbError)
		return
	}
	response.ResponseSuccess(c, data)
}
