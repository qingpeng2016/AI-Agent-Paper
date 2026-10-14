package entity

import (
	"time"

	"gorm.io/datatypes"
)

type PaperManuscriptSubmission struct {
	ID               uint64         `gorm:"primaryKey;column:id"`
	ManuscriptID     uint64         `gorm:"column:manuscript_id;not null"`
	UserID           uint64         `gorm:"column:user_id;not null"`
	VenueCode        string         `gorm:"column:venue_code;size:64;not null"`
	VenueLabel       string         `gorm:"column:venue_label;size:128;not null"`
	Track            string         `gorm:"column:track;size:16;not null;default:main"`
	Round            int            `gorm:"column:round;not null;default:0"`
	Status           string         `gorm:"column:status;size:24;not null;default:submitted"`
	DraftVersion     int            `gorm:"column:draft_version;not null;default:0"`
	Anonymous        bool           `gorm:"column:anonymous;not null;default:1"`
	PortalURL        *string        `gorm:"column:portal_url;size:1024"`
	ExternalID       *string        `gorm:"column:external_id;size:128"`
	Title            *string        `gorm:"column:title;size:512"`
	Abstract         *string        `gorm:"column:abstract"`
	MainPDFURI       *string        `gorm:"column:main_pdf_uri;size:1024"`
	MainPDFFilename  *string        `gorm:"column:main_pdf_filename;size:256"`
	Decision         *string        `gorm:"column:decision;size:128"`
	CommentsSummary  *string        `gorm:"column:comments_summary"`
	RevisionDeadline *time.Time     `gorm:"column:revision_deadline"`
	SubmittedAt      *time.Time     `gorm:"column:submitted_at"`
	Checklist        datatypes.JSON `gorm:"column:checklist"`
	RecordStatus     string         `gorm:"column:record_status;size:16;not null;default:active"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at"`
}

func (PaperManuscriptSubmission) TableName() string { return "paper_manuscript_submission" }

type PaperManuscriptSubmissionEvent struct {
	ID           uint64    `gorm:"primaryKey;column:id"`
	SubmissionID uint64    `gorm:"column:submission_id;not null"`
	EventType    string    `gorm:"column:event_type;size:32;not null"`
	Message      string    `gorm:"column:message;size:512;not null"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (PaperManuscriptSubmissionEvent) TableName() string { return "paper_manuscript_submission_event" }
