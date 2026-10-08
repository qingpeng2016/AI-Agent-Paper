package response

type TopicDiscoveryFormOptions struct {
	Disciplines      []TopicDiscoveryDisciplineOption      `json:"disciplines"`
	ExecutionIntents []TopicDiscoveryExecutionIntensityOption `json:"execution_intensities"`
	AuditLevels      []TopicDiscoveryAuditLevelOption      `json:"audit_levels"`
}

type TopicDiscoveryDisciplineOption struct {
	Code   string   `json:"code"`
	Label  string   `json:"label"`
	NameEn string   `json:"name_en,omitempty"`
	Sort   int      `json:"sort"`
	Sources []string `json:"literature_source_codes,omitempty"`
}

type TopicDiscoveryExecutionIntensityOption struct {
	Code       string  `json:"code"`
	Label      string  `json:"label"`
	Multiplier float64 `json:"multiplier"`
	MaxPapers  uint    `json:"max_papers"`
	MaxIdeas   uint    `json:"max_ideas"`
	IsDefault  bool    `json:"is_default"`
}

type TopicDiscoveryAuditLevelOption struct {
	Code                 string `json:"code"`
	Label                string `json:"label"`
	CitationStrength     uint8  `json:"citation_strength"`
	ClaimStrength        uint8  `json:"claim_strength"`
	KillArgumentStrength uint8  `json:"kill_argument_strength"`
	AuditRounds          uint8  `json:"audit_rounds"`
	IsDefault            bool   `json:"is_default"`
}
