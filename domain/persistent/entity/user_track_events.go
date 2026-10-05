package entity

import "time"

const (
	TrackEventTypePageView = "page_view"
	TrackEventTypeClick    = "click"
	TrackActionEnter       = "enter"
	TrackActionClick       = "click"
)

type UserTrackEvents struct {
	ID           uint      `gorm:"primaryKey;column:id"`
	EventType    string    `gorm:"column:event_type;size:16;not null"`
	Action       string    `gorm:"column:action;size:16;not null"`
	UserID       uint      `gorm:"column:user_id;not null;default:0"`
	VisitorID    string    `gorm:"column:visitor_id;size:64;not null;default:''"`
	SessionID    string    `gorm:"column:session_id;size:64;not null;default:''"`
	Channel      string    `gorm:"column:channel;size:16;not null;default:''"`
	AppVersion   string    `gorm:"column:app_version;size:32;not null;default:''"`
	PageID       string    `gorm:"column:page_id;size:64;not null;default:''"`
	PagePath     string    `gorm:"column:page_path;size:256;not null;default:''"`
	PageTitle    string    `gorm:"column:page_title;size:128;not null;default:''"`
	ElementID    string    `gorm:"column:element_id;size:128;not null;default:''"`
	ElementName  string    `gorm:"column:element_name;size:128;not null;default:''"`
	TargetURL    string    `gorm:"column:target_url;size:512;not null;default:''"`
	APIMethod    string    `gorm:"column:api_method;size:16;not null;default:''"`
	APIPath      string    `gorm:"column:api_path;size:256;not null;default:''"`
	APIParams    *string   `gorm:"column:api_params;type:json"`
	Locale       string    `gorm:"column:locale;size:16;not null;default:zh-Hans"`
	IP           string    `gorm:"column:ip;size:64;not null;default:''"`
	UserAgent    string    `gorm:"column:user_agent;size:512;not null;default:''"`
	DeviceType   string    `gorm:"column:device_type;size:32;not null;default:''"`
	DeviceModel  string    `gorm:"column:device_model;size:128;not null;default:''"`
	OSName       string    `gorm:"column:os_name;size:32;not null;default:''"`
	OSVersion    string    `gorm:"column:os_version;size:32;not null;default:''"`
	ScreenWidth  int       `gorm:"column:screen_width;not null;default:0"`
	ScreenHeight int       `gorm:"column:screen_height;not null;default:0"`
	Referrer     string    `gorm:"column:referrer;size:512;not null;default:''"`
	ExtraJSON    *string   `gorm:"column:extra_json;type:json"`
	EventAt      time.Time `gorm:"column:event_at;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (UserTrackEvents) TableName() string { return "user_track_events" }
