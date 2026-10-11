package entity

import (
	"time"

	"gorm.io/datatypes"
)

type PaperRefVenue struct {
	ID                  uint           `gorm:"primaryKey;column:id"`
	Code                string         `gorm:"column:code;size:64;not null"`
	Name                string         `gorm:"column:name;size:256;not null"`
	VenueType           string         `gorm:"column:venue_type;size:16;not null"`
	LLMDescriptionCodes datatypes.JSON `gorm:"column:llm_description_codes;not null"`
	Publisher           *string        `gorm:"column:publisher;size:128"`
	WebsiteURL          *string        `gorm:"column:website_url;size:512"`
	Status              string         `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt           time.Time      `gorm:"column:created_at"`
	UpdatedAt           time.Time      `gorm:"column:updated_at"`
}

func (PaperRefVenue) TableName() string { return "paper_ref_venue" }
