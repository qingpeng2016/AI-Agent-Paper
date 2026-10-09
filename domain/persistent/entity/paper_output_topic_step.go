package entity

import (
	"time"

	"gorm.io/datatypes"
)

type PaperOutputTopicStep struct {
	ID            uint64         `gorm:"primaryKey;column:id"`
	ManuscriptID  uint64         `gorm:"column:manuscript_id;not null"`
	UserID        uint64         `gorm:"column:user_id;not null"`
	RunVersion    int            `gorm:"column:run_version;not null;default:1"`
	StageCode     string         `gorm:"column:stage_code;size:32;not null"`
	Status        string         `gorm:"column:status;size:16;not null;default:pending"`
	Result        datatypes.JSON `gorm:"column:result"`
	SummaryText   *string        `gorm:"column:summary_text"`
	InputParams   datatypes.JSON `gorm:"column:input_params"`
	Extra         datatypes.JSON `gorm:"column:extra"`
	StartedAt     *time.Time     `gorm:"column:started_at"`
	CompletedAt   *time.Time     `gorm:"column:completed_at"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
}

func (PaperOutputTopicStep) TableName() string { return "paper_output_topic_step" }
