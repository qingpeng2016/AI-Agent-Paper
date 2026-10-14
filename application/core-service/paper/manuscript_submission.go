package paper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/request"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
	"gorm.io/datatypes"
)

const maxSubmissionPDFBytes = 30 << 20 // 30MB

var revisionDecisionPattern = regexp.MustCompile(`(?i)major|minor|revision`)

type ManuscriptSubmissionService struct {
	submissions repository.PaperManuscriptSubmissionRepo
	manuscripts repository.PaperManuscriptRepo
}

func NewManuscriptSubmissionService(
	submissions repository.PaperManuscriptSubmissionRepo,
	manuscripts repository.PaperManuscriptRepo,
) *ManuscriptSubmissionService {
	return &ManuscriptSubmissionService{submissions: submissions, manuscripts: manuscripts}
}

func (s *ManuscriptSubmissionService) ListByManuscript(
	ctx context.Context,
	userID uint,
	manuscriptID uint64,
	page, pageSize int,
) (*response.PaperManuscriptSubmissionListView, error) {
	if _, err := s.ensureManuscript(ctx, userID, manuscriptID); err != nil {
		return nil, err
	}
	page, pageSize, offset := NormalizeListPage(page, pageSize)
	total, err := s.submissions.CountByManuscript(ctx, manuscriptID)
	if err != nil {
		return nil, err
	}
	rows, err := s.submissions.ListByManuscriptPaged(ctx, manuscriptID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	out := &response.PaperManuscriptSubmissionListView{
		ManuscriptID: manuscriptID,
		Items:        make([]response.PaperManuscriptSubmissionItemView, 0, len(rows)),
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
	}
	for _, row := range rows {
		out.Items = append(out.Items, submissionItemView(row))
	}
	return out, nil
}

func (s *ManuscriptSubmissionService) GetDetail(
	ctx context.Context,
	userID uint,
	manuscriptID, submissionID uint64,
) (*response.PaperManuscriptSubmissionDetailView, error) {
	if _, err := s.ensureManuscript(ctx, userID, manuscriptID); err != nil {
		return nil, err
	}
	row, err := s.submissions.GetByIDForUser(ctx, submissionID, manuscriptID, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errorx.ErrParamsError.WithDetail("投递记录不存在或无权访问")
	}
	events, err := s.submissions.ListEvents(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	detail := &response.PaperManuscriptSubmissionDetailView{
		PaperManuscriptSubmissionItemView: submissionItemView(*row),
		Checklist:                         decodeChecklist(row.Checklist),
		Events:                            make([]response.PaperManuscriptSubmissionEventView, 0, len(events)),
	}
	for _, ev := range events {
		detail.Events = append(detail.Events, response.PaperManuscriptSubmissionEventView{
			ID:        strconv.FormatUint(ev.ID, 10),
			EventType: ev.EventType,
			Message:   ev.Message,
			CreatedAt: ev.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return detail, nil
}

func (s *ManuscriptSubmissionService) Create(
	ctx context.Context,
	userID uint,
	body request.PaperManuscriptSubmissionCreateBody,
) (*response.PaperManuscriptSubmissionDetailView, error) {
	if !body.MarkedSubmitted {
		return nil, errorx.ErrParamsError.WithDetail("请确认已在投稿系统完成提交")
	}
	ms, err := s.ensureManuscript(ctx, userID, body.ManuscriptID)
	if err != nil {
		return nil, err
	}
	if body.DraftVersion < 1 {
		return nil, errorx.ErrParamsError.WithDetail("请选择有效稿本版本")
	}
	if ms.DraftVersion > 0 && body.DraftVersion > int(ms.DraftVersion) {
		return nil, errorx.ErrParamsError.WithDetail("稿本版本超出当前写作版本")
	}
	venueCode := strings.TrimSpace(body.VenueCode)
	venueLabel := strings.TrimSpace(body.VenueLabel)
	if venueCode == "" || venueLabel == "" {
		return nil, errorx.ErrParamsError.WithDetail("请选择目标 venue")
	}
	track := normalizeSubmissionTrack(body.Track)
	now := time.Now()
	checklist := defaultSubmissionChecklistJSON()
	row := &entity.PaperManuscriptSubmission{
		ManuscriptID: body.ManuscriptID,
		UserID:       uint64(userID),
		VenueCode:    venueCode,
		VenueLabel:   venueLabel,
		Track:        track,
		Round:        0,
		Status:       "submitted",
		DraftVersion: body.DraftVersion,
		Anonymous:    body.Anonymous,
		Checklist:    checklist,
		SubmittedAt:  &now,
	}
	if v := strings.TrimSpace(body.PortalURL); v != "" {
		row.PortalURL = &v
	}
	if v := strings.TrimSpace(body.ExternalID); v != "" {
		row.ExternalID = &v
	}
	if v := strings.TrimSpace(body.Title); v != "" {
		row.Title = &v
	}
	if v := strings.TrimSpace(body.Abstract); v != "" {
		row.Abstract = &v
	}
	if err := s.submissions.Create(ctx, row); err != nil {
		return nil, err
	}
	_ = s.submissions.InsertEvent(ctx, &entity.PaperManuscriptSubmissionEvent{
		SubmissionID: row.ID,
		EventType:    "created",
		Message:      "创建投递记录",
	})
	_ = s.submissions.InsertEvent(ctx, &entity.PaperManuscriptSubmissionEvent{
		SubmissionID: row.ID,
		EventType:    "submitted",
		Message:      "标记已在投稿系统完成提交",
	})
	if err := s.syncManuscriptSummary(ctx, userID, body.ManuscriptID); err != nil {
		return nil, err
	}
	return s.GetDetail(ctx, userID, body.ManuscriptID, row.ID)
}

func (s *ManuscriptSubmissionService) UploadMainPDF(
	ctx context.Context,
	userID uint,
	manuscriptID, submissionID uint64,
	originalFileName string,
	content io.Reader,
) (*response.PaperManuscriptSubmissionDetailView, error) {
	if _, err := s.ensureManuscript(ctx, userID, manuscriptID); err != nil {
		return nil, err
	}
	row, err := s.submissions.GetByIDForUser(ctx, submissionID, manuscriptID, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errorx.ErrParamsError.WithDetail("投递记录不存在或无权访问")
	}
	rel := SubmissionMainPDFLocalRelPath(manuscriptID, submissionID, originalFileName)
	rel = filepath.ToSlash(rel)
	abs := StorageLocalAbsPath(rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return nil, err
	}
	tmp := abs + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return nil, err
	}
	n, copyErr := io.Copy(f, io.LimitReader(content, maxSubmissionPDFBytes+1))
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return nil, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return nil, closeErr
	}
	if n > maxSubmissionPDFBytes {
		_ = os.Remove(tmp)
		return nil, errorx.ErrParamsError.WithDetail(fmt.Sprintf("PDF 超过 %dMB 限制", maxSubmissionPDFBytes>>20))
	}
	if n == 0 {
		_ = os.Remove(tmp)
		return nil, errorx.ErrParamsError.WithDetail("文件为空")
	}
	if err := os.Rename(tmp, abs); err != nil {
		_ = os.Remove(tmp)
		return nil, err
	}
	if row.MainPDFURI != nil && strings.TrimSpace(*row.MainPDFURI) != "" {
		_ = os.Remove(StorageLocalAbsPath(*row.MainPDFURI))
	}
	fname := filepath.Base(originalFileName)
	ok, err := s.submissions.UpdateMainPDF(ctx, submissionID, manuscriptID, userID, rel, fname)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errorx.ErrParamsError.WithDetail("更新 PDF 失败")
	}
	_ = s.submissions.InsertEvent(ctx, &entity.PaperManuscriptSubmissionEvent{
		SubmissionID: submissionID,
		EventType:    "pdf_uploaded",
		Message:      "上传主 PDF：" + fname,
	})
	return s.GetDetail(ctx, userID, manuscriptID, submissionID)
}

func (s *ManuscriptSubmissionService) UpdateRevision(
	ctx context.Context,
	userID uint,
	body request.PaperManuscriptSubmissionRevisionBody,
) (*response.PaperManuscriptSubmissionDetailView, error) {
	if _, err := s.ensureManuscript(ctx, userID, body.ManuscriptID); err != nil {
		return nil, err
	}
	row, err := s.submissions.GetByIDForUser(ctx, body.SubmissionID, body.ManuscriptID, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errorx.ErrParamsError.WithDetail("投递记录不存在或无权访问")
	}
	decision := strings.TrimSpace(body.Decision)
	comments := strings.TrimSpace(body.Comments)
	status := row.Status
	var deadline *time.Time
	if revisionDecisionPattern.MatchString(decision) {
		status = "revision"
		t := time.Now().Add(45 * 24 * time.Hour)
		deadline = &t
	}
	decPtr, comPtr := &decision, &comments
	if decision == "" {
		decPtr = nil
	}
	if comments == "" {
		comPtr = nil
	}
	ok, err := s.submissions.UpdateRevision(ctx, body.SubmissionID, body.ManuscriptID, userID, status, decPtr, comPtr, deadline)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errorx.ErrParamsError.WithDetail("保存改稿计划失败")
	}
	_ = s.submissions.InsertEvent(ctx, &entity.PaperManuscriptSubmissionEvent{
		SubmissionID: body.SubmissionID,
		EventType:    "revision_saved",
		Message:      "保存改稿计划",
	})
	if err := s.syncManuscriptSummary(ctx, userID, body.ManuscriptID); err != nil {
		return nil, err
	}
	return s.GetDetail(ctx, userID, body.ManuscriptID, body.SubmissionID)
}

func (s *ManuscriptSubmissionService) MarkResubmitted(
	ctx context.Context,
	userID uint,
	manuscriptID, submissionID uint64,
) (*response.PaperManuscriptSubmissionDetailView, error) {
	if _, err := s.ensureManuscript(ctx, userID, manuscriptID); err != nil {
		return nil, err
	}
	row, err := s.submissions.GetByIDForUser(ctx, submissionID, manuscriptID, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errorx.ErrParamsError.WithDetail("投递记录不存在或无权访问")
	}
	newRound := row.Round + 1
	ok, err := s.submissions.MarkResubmitted(ctx, submissionID, manuscriptID, userID, newRound, "resubmitted")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errorx.ErrParamsError.WithDetail("标记修回失败")
	}
	_ = s.submissions.InsertEvent(ctx, &entity.PaperManuscriptSubmissionEvent{
		SubmissionID: submissionID,
		EventType:    "resubmitted",
		Message:      fmt.Sprintf("标记已修回 · R%d", newRound),
	})
	if err := s.syncManuscriptSummary(ctx, userID, manuscriptID); err != nil {
		return nil, err
	}
	return s.GetDetail(ctx, userID, manuscriptID, submissionID)
}

func (s *ManuscriptSubmissionService) ensureManuscript(
	ctx context.Context,
	userID uint,
	manuscriptID uint64,
) (*entity.PaperManuscript, error) {
	if manuscriptID == 0 {
		return nil, errorx.ErrParamsError
	}
	ms, err := s.manuscripts.GetByIDForUser(ctx, uint(manuscriptID), userID)
	if err != nil {
		return nil, err
	}
	if ms == nil {
		return nil, errorx.ErrParamsError.WithDetail("manuscript 不存在或无权访问")
	}
	return ms, nil
}

func (s *ManuscriptSubmissionService) syncManuscriptSummary(ctx context.Context, userID uint, manuscriptID uint64) error {
	latest, err := s.submissions.GetLatestForManuscript(ctx, manuscriptID)
	if err != nil {
		return err
	}
	if latest == nil {
		return s.manuscripts.UpdateSubmissionSummary(ctx, uint(manuscriptID), userID, "not_submitted", nil, nil)
	}
	label := submissionSummaryLabel(*latest)
	var deadline *time.Time
	if latest.RevisionDeadline != nil && !latest.RevisionDeadline.IsZero() {
		deadline = latest.RevisionDeadline
	}
	return s.manuscripts.UpdateSubmissionSummary(ctx, uint(manuscriptID), userID, latest.Status, &label, deadline)
}

func submissionSummaryLabel(row entity.PaperManuscriptSubmission) string {
	track := "Main"
	if strings.EqualFold(row.Track, "short") {
		track = "Short"
	}
	return fmt.Sprintf("%s · %s", row.VenueLabel, track)
}

func normalizeSubmissionTrack(track string) string {
	t := strings.ToLower(strings.TrimSpace(track))
	if t == "short" {
		return "short"
	}
	return "main"
}

func submissionItemView(row entity.PaperManuscriptSubmission) response.PaperManuscriptSubmissionItemView {
	item := response.PaperManuscriptSubmissionItemView{
		ID:           strconv.FormatUint(row.ID, 10),
		VenueCode:    row.VenueCode,
		VenueLabel:   row.VenueLabel,
		Track:        row.Track,
		Round:        row.Round,
		Status:       row.Status,
		DraftVersion: row.DraftVersion,
		Anonymous:    row.Anonymous,
		CreatedAt:    row.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    row.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if row.PortalURL != nil {
		item.PortalURL = strings.TrimSpace(*row.PortalURL)
	}
	if row.ExternalID != nil {
		item.ExternalID = strings.TrimSpace(*row.ExternalID)
	}
	if row.Title != nil {
		item.Title = strings.TrimSpace(*row.Title)
	}
	if row.Abstract != nil {
		item.Abstract = strings.TrimSpace(*row.Abstract)
	}
	if row.MainPDFFilename != nil {
		item.MainPDFFilename = strings.TrimSpace(*row.MainPDFFilename)
	}
	if row.Decision != nil {
		item.Decision = strings.TrimSpace(*row.Decision)
	}
	if row.CommentsSummary != nil {
		item.CommentsSummary = strings.TrimSpace(*row.CommentsSummary)
	}
	if row.RevisionDeadline != nil && !row.RevisionDeadline.IsZero() {
		item.RevisionDeadline = row.RevisionDeadline.UTC().Format(time.RFC3339)
	}
	if row.SubmittedAt != nil && !row.SubmittedAt.IsZero() {
		item.SubmittedAt = row.SubmittedAt.UTC().Format(time.RFC3339)
	}
	return item
}

type checklistItem struct {
	ID string `json:"id"`
	OK bool   `json:"ok"`
}

func defaultSubmissionChecklistJSON() datatypes.JSON {
	raw, _ := json.Marshal([]checklistItem{
		{ID: "pages", OK: true},
		{ID: "anon", OK: true},
		{ID: "supp", OK: false},
	})
	return datatypes.JSON(raw)
}

func decodeChecklist(raw datatypes.JSON) []response.PaperManuscriptSubmissionChecklistItemView {
	if len(raw) == 0 {
		return nil
	}
	var items []checklistItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	labels := map[string]string{
		"pages": "页数符合 venue 限制",
		"anon":  "PDF 匿名化检查",
		"supp":  "补充材料已打包",
	}
	out := make([]response.PaperManuscriptSubmissionChecklistItemView, 0, len(items))
	for _, it := range items {
		out = append(out, response.PaperManuscriptSubmissionChecklistItemView{
			ID:    it.ID,
			OK:    it.OK,
			Label: labels[it.ID],
		})
	}
	return out
}
