package request

type TopicDiscoveryRunRequest struct {
	ManuscriptID    uint64   `json:"manuscript_id"`
	ManuscriptTitle string   `json:"manuscript_title"`
	DisciplineCode  string   `json:"discipline_code"`
	Direction       string   `json:"direction"`
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
