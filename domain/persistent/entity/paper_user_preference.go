package entity

import (
	"time"

	"gorm.io/datatypes"
)

type PaperUserPreference struct {
	UserID                   uint64         `gorm:"primaryKey;column:user_id"`
	DefaultDisciplineID      *uint64        `gorm:"column:default_discipline_id"`
	DefaultVenueID           *uint64        `gorm:"column:default_venue_id"`
	DefaultIntensityCode     string         `gorm:"column:default_intensity_code;size:16;not null;default:balanced"`
	DefaultAuditLevelCode    string         `gorm:"column:default_audit_level_code;size:16;not null;default:standard"`
	DefaultHumanCheckpoint   bool           `gorm:"column:default_human_checkpoint;not null;default:1"`
	DefaultLiteratureSourceIDs datatypes.JSON `gorm:"column:default_literature_source_ids;type:json"`
	UpdatedAt                time.Time      `gorm:"column:updated_at"`
}

func (PaperUserPreference) TableName() string { return "paper_user_preference" }
