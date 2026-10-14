package entity

const (
	LLMSceneTopic      = "topic"
	LLMSceneWritePaper = "write_paper"
)

type PaperLLMConstraintGate struct {
	PromptScene string `gorm:"primaryKey;column:prompt_scene;size:32"`
	Code        string `gorm:"primaryKey;column:code;size:16"`
	Name        string `gorm:"column:name;size:64;not null"`
	IsDefault   bool   `gorm:"column:is_default;not null;default:0"`
}

func (PaperLLMConstraintGate) TableName() string { return "paper_llm_constraint_gate" }
