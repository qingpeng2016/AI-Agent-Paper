package request

type TopicDiscoveryRunRequest struct {
	ManuscriptID    uint64   `json:"manuscript_id"`
	ManuscriptTitle string   `json:"manuscript_title"`
	DisciplineCode  string   `json:"discipline_code"`
	Keywords        []string `json:"keywords"`
	Description     string   `json:"description"`
	Direction       string   `json:"direction"` // 兼容旧版单框；新版请传 keywords + description
	Venue           string   `json:"venue"`
	VenueCodes      []string `json:"venue_codes"`
	SourceCodes     []string `json:"source_codes"`
	Intensity       string   `json:"intensity"`
	AuditLevel      string   `json:"audit_level"`
	HumanCheckpoint   bool `json:"human_checkpoint"`
	RandomReferences   bool     `json:"random_references"`
	UserLiteratureIDs  []string `json:"user_literature_ids"`
	// start | continue | resume | run_all — resume：原生 run_state 续跑 idea_discovery（超时/CLI 中断后）
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

type TopicDiscoveryConfirmRoundRequest struct {
	ManuscriptID  uint64 `json:"manuscript_id" binding:"required"`
	RunVersion    int    `json:"run_version"`
	SelectedRound int    `json:"selected_round" binding:"required"`
}
