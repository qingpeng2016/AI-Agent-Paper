package entity

type PaperExecIntensityTopic struct {
	ID         uint   `gorm:"primaryKey;column:id"`
	Code       string `gorm:"column:code;size:16;not null;uniqueIndex:uk_paper_exec_intensity_topic_code"`
	Name       string `gorm:"column:name;size:64;not null"`
	ExecRounds uint8  `gorm:"column:exec_rounds;not null;default:1"`
	MaxPapers  uint   `gorm:"column:max_papers;not null"`
	MaxIdeas   uint   `gorm:"column:max_ideas;not null"`
	IsDefault  bool   `gorm:"column:is_default;not null;default:0"`
}

func (PaperExecIntensityTopic) TableName() string {
	return "paper_exec_intensity_topic"
}
