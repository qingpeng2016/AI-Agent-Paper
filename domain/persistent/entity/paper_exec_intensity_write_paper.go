package entity

type PaperExecIntensityWritePaper struct {
	Code       string `gorm:"primaryKey;column:code;size:16"`
	Name       string `gorm:"column:name;size:64;not null"`
	ExecRounds uint8  `gorm:"column:exec_rounds;not null;default:1"`
	IsDefault  bool   `gorm:"column:is_default;not null;default:0"`
}

func (PaperExecIntensityWritePaper) TableName() string {
	return "paper_exec_intensity_write_paper"
}
