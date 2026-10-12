package response

type TopicDiscoveryFormOptions struct {
	Disciplines       []TopicDiscoveryDisciplineOption           `json:"disciplines"`
	Venues            []TopicDiscoveryVenueOption                `json:"venues"`
	ExecutionIntents  []TopicDiscoveryExecutionIntensityOption   `json:"execution_intensities"`
	AuditLevels       []TopicDiscoveryAuditLevelOption           `json:"audit_levels"`
	LiteratureSources []TopicDiscoveryLiteratureSourceOption     `json:"literature_sources"`
}

type TopicDiscoveryVenueOption struct {
	Code      string `json:"code"`
	Label     string `json:"label"`
	VenueType string `json:"venue_type"`
	Sort      int    `json:"sort"`
}

type TopicDiscoveryLiteratureSourceOption struct {
	Code            string `json:"code"`
	Label           string `json:"label"`
	Sort            int    `json:"sort"`
	DefaultSelected bool   `json:"default_selected"`
}

type TopicDiscoveryDisciplineOption struct {
	Code   string   `json:"code"`
	Label  string   `json:"label"`
	NameEn string   `json:"name_en,omitempty"`
	Sort int `json:"sort"`
}

type TopicDiscoveryExecutionIntensityOption struct {
	Code       string  `json:"code"`
	Label      string  `json:"label"`
	ExecRounds uint8  `json:"exec_rounds"`
	MaxPapers  uint    `json:"max_papers"`
	MaxIdeas   uint    `json:"max_ideas"`
	IsDefault  bool    `json:"is_default"`
}

type TopicDiscoveryAuditLevelOption struct {
	Code            string `json:"code"`
	Label           string `json:"label"`
	LlmDescriptionCitationCheckCode uint64 `json:"llm_description_citation_check_code"`
	LlmDescriptionClaimCheckCode    uint64 `json:"llm_description_claim_check_code"`
	LlmDescriptionKillCheckCode     uint64 `json:"llm_description_kill_check_code"`
	IsDefault       bool   `json:"is_default"`
}
