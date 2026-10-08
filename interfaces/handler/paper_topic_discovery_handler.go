package handler

import (
	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/response"
	"github.com/gin-gonic/gin"
)

type PaperTopicDiscoveryHandler struct {
	opts *papersvc.TopicDiscoveryOptionsService
}

func NewPaperTopicDiscoveryHandler(opts *papersvc.TopicDiscoveryOptionsService) *PaperTopicDiscoveryHandler {
	return &PaperTopicDiscoveryHandler{opts: opts}
}

// GetFormOptions GET /api/v1/paper/topic-discovery/form-options
func (h *PaperTopicDiscoveryHandler) GetFormOptions(c *gin.Context) {
	data, err := h.opts.GetFormOptions(c.Request.Context())
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
