package entity

type PaperRefAuditLevel struct {
	Code                 string `gorm:"primaryKey;column:code;size:16"`
	Name                 string `gorm:"column:name;size:64;not null"`
	CitationStrength     uint8  `gorm:"column:citation_strength;not null"`
	ClaimStrength        uint8  `gorm:"column:claim_strength;not null"`
	KillArgumentStrength uint8  `gorm:"column:kill_argument_strength;not null"`
	AuditRounds          uint8  `gorm:"column:audit_rounds;not null"`
	IsDefault            bool   `gorm:"column:is_default;not null;default:0"`
}

func (PaperRefAuditLevel) TableName() string { return "paper_ref_audit_level" }
