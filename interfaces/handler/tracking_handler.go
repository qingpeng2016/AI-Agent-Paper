package handler

import (
	trackingSvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/tracking"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/request"
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
