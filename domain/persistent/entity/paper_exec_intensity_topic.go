package entity

type PaperExecIntensityTopic struct {
	Code       string `gorm:"primaryKey;column:code;size:16"`
	Name       string `gorm:"column:name;size:64;not null"`
	ExecRounds uint8  `gorm:"column:exec_rounds;not null;default:1"`
	MaxPapers  uint   `gorm:"column:max_papers;not null"`
	MaxIdeas   uint   `gorm:"column:max_ideas;not null"`
	IsDefault  bool   `gorm:"column:is_default;not null;default:0"`
}

func (PaperExecIntensityTopic) TableName() string {
	return "paper_exec_intensity_topic"
}
