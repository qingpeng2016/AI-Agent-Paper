package entity

import "time"

// 用户文献 `from` 与 paper_literature_platform.code 对齐；用户上传用 user_upload。
const (
	UserLiteratureFromUserUpload       = "user_upload"
	UserLiteratureFromArxiv             = "arxiv"
	UserLiteratureFromOpenAlex          = "openalex"
	UserLiteratureFromSemanticScholar   = "semantic_scholar"
	UserLiteratureFromCrossref          = "crossref"
	UserLiteratureFromPubMed            = "pubmed"
)

type PaperUserLiterature struct {
	ID          uint64    `gorm:"primaryKey;column:id"`
	UserID      uint64    `gorm:"column:user_id;not null"`
	From        string    `gorm:"column:from;size:64;not null;default:user_upload"`
	FileExt     string    `gorm:"column:file_ext;size:16;not null"`
	FileName    string    `gorm:"column:file_name;size:512;not null"`
	LiteratureTitle string    `gorm:"column:literature_title;size:512;not null"`
	Note        *string   `gorm:"column:note"`
	SizeBytes   uint64    `gorm:"column:size_bytes;not null;default:0"`
	StorageURL  *string   `gorm:"column:storage_url;size:1024"`
	ContentHash *string   `gorm:"column:content_hash;size:64"`
	Status      string    `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (PaperUserLiterature) TableName() string { return "paper_user_literature" }
