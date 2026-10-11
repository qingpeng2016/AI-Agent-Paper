package entity

type PaperLLMConstraintExecIntensity struct {
	Code       string `gorm:"primaryKey;column:code;size:16"`
	Name       string `gorm:"column:name;size:64;not null"`
	ExecRounds uint8  `gorm:"column:exec_rounds;not null;default:1"`
	MaxPapers  uint   `gorm:"column:max_papers;not null"`
	MaxIdeas   uint   `gorm:"column:max_ideas;not null"`
	IsDefault  bool   `gorm:"column:is_default;not null;default:0"`
}

func (PaperLLMConstraintExecIntensity) TableName() string {
	return "paper_llm_constraint_exec_intensity"
}
