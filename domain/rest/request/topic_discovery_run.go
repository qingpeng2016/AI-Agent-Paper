package request

type TopicDiscoveryRunRequest struct {
	ManuscriptID    uint64   `json:"manuscript_id"`
	ManuscriptTitle string   `json:"manuscript_title"`
	DisciplineCode  string   `json:"discipline_code"`
	Keywords        []string `json:"keywords"`
	Description     string   `json:"description"`
	Direction       string   `json:"direction"` // 兼容旧版单框；新版请传 keywords + description
	Venue           string   `json:"venue"`
	SourceCodes     []string `json:"source_codes"`
	Intensity       string   `json:"intensity"`
	AuditLevel      string   `json:"audit_level"`
	HumanCheckpoint bool     `json:"human_checkpoint"`
	// start：新开 run 并执行第一步；continue：执行下一个 pending 步；run_all：连续执行直到完成或需人工检查点
	Action string `json:"action"`
}

type TopicDiscoveryRunQuery struct {
	ManuscriptID uint64 `form:"manuscript_id" binding:"required"`
}

type TopicDiscoveryCurrentRunQuery struct {
	ManuscriptID uint64 `form:"manuscript_id" binding:"required"`
}

type TopicDiscoveryCancelRequest struct {
	ManuscriptID uint64 `json:"manuscript_id"`
}

type TopicDiscoveryCommitManuscriptRequest struct {
	ManuscriptTitle string `json:"manuscript_title"`
}
