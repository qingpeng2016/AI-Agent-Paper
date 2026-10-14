package handler

import (
	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type PaperManuscriptSubmissionHandler struct {
	svc *papersvc.ManuscriptSubmissionService
}

func NewPaperManuscriptSubmissionHandler(svc *papersvc.ManuscriptSubmissionService) *PaperManuscriptSubmissionHandler {
	return &PaperManuscriptSubmissionHandler{svc: svc}
}

// GetManuscriptSubmissions GET /api/v1/paper/manuscript-submissions
func (h *PaperManuscriptSubmissionHandler) GetManuscriptSubmissions(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperManuscriptSubmissionListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.ListByManuscript(c.Request.Context(), userID, q.ManuscriptID, q.Page, q.PageSize)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// GetManuscriptSubmissionDetail GET /api/v1/paper/manuscript-submissions/detail
func (h *PaperManuscriptSubmissionHandler) GetManuscriptSubmissionDetail(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperManuscriptSubmissionDetailQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.GetDetail(c.Request.Context(), userID, q.ManuscriptID, q.SubmissionID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PostCreateManuscriptSubmission POST /api/v1/paper/manuscript-submissions
func (h *PaperManuscriptSubmissionHandler) PostCreateManuscriptSubmission(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var body request.PaperManuscriptSubmissionCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.Create(c.Request.Context(), userID, body)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PutManuscriptSubmissionRevision PUT /api/v1/paper/manuscript-submissions/revision
func (h *PaperManuscriptSubmissionHandler) PutManuscriptSubmissionRevision(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var body request.PaperManuscriptSubmissionRevisionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.UpdateRevision(c.Request.Context(), userID, body)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PostMarkManuscriptSubmissionResubmitted POST /api/v1/paper/manuscript-submissions/mark-resubmitted
func (h *PaperManuscriptSubmissionHandler) PostMarkManuscriptSubmissionResubmitted(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var body request.PaperManuscriptSubmissionMarkResubmittedBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	data, err := h.svc.MarkResubmitted(c.Request.Context(), userID, body.ManuscriptID, body.SubmissionID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PostUploadManuscriptSubmissionPDF POST /api/v1/paper/manuscript-submissions/upload-main-pdf
func (h *PaperManuscriptSubmissionHandler) PostUploadManuscriptSubmissionPDF(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	var q request.PaperManuscriptSubmissionUploadForm
	if err := c.ShouldBind(&q); err != nil {
		response.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseErr(c, errorx.ErrParamsError.WithDetail("请选择 PDF 文件"))
		return
	}
	f, err := file.Open()
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	defer f.Close()
	data, err := h.svc.UploadMainPDF(
		c.Request.Context(),
		userID,
		q.ManuscriptID,
		q.SubmissionID,
		file.Filename,
		f,
	)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
