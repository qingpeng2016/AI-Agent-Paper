package entity

type PaperLLMConstraintGate struct {
	Code                          string `gorm:"primaryKey;column:code;size:16"`
	Name                          string `gorm:"column:name;size:64;not null"`
	LlmDescriptionCitationCheckCode uint64 `gorm:"column:llm_description_citation_check_code;not null;default:0"`
	LlmDescriptionClaimCheckCode    uint64 `gorm:"column:llm_description_claim_check_code;not null;default:0"`
	LlmDescriptionKillCheckCode     uint64 `gorm:"column:llm_description_kill_check_code;not null;default:0"`
	IsDefault                     bool   `gorm:"column:is_default;not null;default:0"`
}

func (PaperLLMConstraintGate) TableName() string { return "paper_llm_constraint_gate" }
