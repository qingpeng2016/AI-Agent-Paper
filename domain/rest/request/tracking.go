package request

import (
	"encoding/json"
	"time"
)

type TrackEventItem struct {
	EventType    string          `json:"event_type" binding:"required"`
	Action       string          `json:"action" binding:"required"`
	VisitorID    string          `json:"visitor_id"`
	SessionID    string          `json:"session_id"`
	Channel      string          `json:"channel"`
	AppVersion   string          `json:"app_version"`
	PageID       string          `json:"page_id"`
	PagePath     string          `json:"page_path"`
	PageTitle    string          `json:"page_title"`
	ElementID    string          `json:"element_id"`
	ElementName  string          `json:"element_name"`
	TargetURL    string          `json:"target_url"`
	APIMethod    string          `json:"api_method"`
	APIPath      string          `json:"api_path"`
	APIParams    json.RawMessage `json:"api_params"`
	Locale       string          `json:"locale"`
	DeviceType   string          `json:"device_type"`
	DeviceModel  string          `json:"device_model"`
	OSName       string          `json:"os_name"`
	OSVersion    string          `json:"os_version"`
	ScreenWidth  int             `json:"screen_width"`
	ScreenHeight int             `json:"screen_height"`
	Referrer     string          `json:"referrer"`
	Extra        json.RawMessage `json:"extra"`
	EventAt      *time.Time      `json:"event_at"`
}

type ReportTrackEventsReq struct {
	Events []TrackEventItem `json:"events" binding:"required,min=1,max=50,dive"`
}
