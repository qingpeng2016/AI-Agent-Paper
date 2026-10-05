package tracking

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	ginMiddleware "github.com/qingpeng2016/ai-token-mall/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
	"github.com/gin-gonic/gin"
)

const maxBatchSize = 50

type Service struct {
	trackRepo repository.UserTrackEventsRepo
}

func NewService(trackRepo repository.UserTrackEventsRepo) *Service {
	return &Service{trackRepo: trackRepo}
}

func (s *Service) ReportEvents(c *gin.Context, req *request.ReportTrackEventsReq) (*response.ReportTrackEventsResp, error) {
	if req == nil || len(req.Events) == 0 {
		return &response.ReportTrackEventsResp{}, nil
	}
	events := req.Events
	if len(events) > maxBatchSize {
		events = events[:maxBatchSize]
	}

	userID, _ := ginMiddleware.TryUserIDFromRequest(c)
	ip := clientIP(c)
	ua := truncate(c.GetHeader("User-Agent"), 512)
	now := time.Now()

	rows := make([]entity.UserTrackEvents, 0, len(events))
	for _, ev := range events {
		eventType := strings.ToLower(strings.TrimSpace(ev.EventType))
		action := strings.ToLower(strings.TrimSpace(ev.Action))
		if !validEventType(eventType) || !validAction(action) {
			continue
		}
		eventAt := now
		if ev.EventAt != nil && !ev.EventAt.IsZero() {
			eventAt = ev.EventAt.UTC()
		}
		rows = append(rows, entity.UserTrackEvents{
			EventType:    eventType,
			Action:       action,
			UserID:       userID,
			VisitorID:    truncate(ev.VisitorID, 64),
			SessionID:    truncate(ev.SessionID, 64),
			Channel:      truncate(strings.ToLower(ev.Channel), 16),
			AppVersion:   truncate(ev.AppVersion, 32),
			PageID:       truncate(ev.PageID, 64),
			PagePath:     truncate(ev.PagePath, 256),
			PageTitle:    truncate(ev.PageTitle, 128),
			ElementID:    truncate(ev.ElementID, 128),
			ElementName:  truncate(ev.ElementName, 128),
			TargetURL:    truncate(ev.TargetURL, 512),
			APIMethod:    truncate(strings.ToUpper(ev.APIMethod), 16),
			APIPath:      truncate(ev.APIPath, 256),
			APIParams:    jsonColumn(ev.APIParams),
			Locale:       truncate(ev.Locale, 16),
			IP:           ip,
			UserAgent:    ua,
			DeviceType:   truncate(ev.DeviceType, 32),
			DeviceModel:  truncate(ev.DeviceModel, 128),
			OSName:       truncate(ev.OSName, 32),
			OSVersion:    truncate(ev.OSVersion, 32),
			ScreenWidth:  ev.ScreenWidth,
			ScreenHeight: ev.ScreenHeight,
			Referrer:     truncate(ev.Referrer, 512),
			ExtraJSON:    jsonColumn(ev.Extra),
			EventAt:      eventAt,
			CreatedAt:    now,
		})
	}
	if len(rows) == 0 {
		return &response.ReportTrackEventsResp{}, nil
	}
	if err := s.trackRepo.CreateBatch(context.Background(), nil, rows); err != nil {
		return nil, err
	}
	return &response.ReportTrackEventsResp{Accepted: len(rows)}, nil
}

func validEventType(v string) bool {
	return v == entity.TrackEventTypePageView || v == entity.TrackEventTypeClick
}

func validAction(v string) bool {
	return v == entity.TrackActionEnter || v == entity.TrackActionClick
}

func jsonColumn(raw json.RawMessage) *string {
	if len(raw) == 0 || !json.Valid(raw) {
		return nil
	}
	s := string(raw)
	return &s
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max]
}

func clientIP(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := c.GetHeader("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	return c.ClientIP()
}
