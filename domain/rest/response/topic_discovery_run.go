package response

import "encoding/json"

type TopicDiscoveryStepView struct {
	StageCode   string          `json:"stage_code"`
	Status      string          `json:"status"`
	SummaryText string          `json:"summary_text,omitempty"`
	InputParams json.RawMessage `json:"input_params,omitempty"`
	Result      json.RawMessage `json:"result,omitempty"`
	Extra       json.RawMessage `json:"extra,omitempty"`
	StartedAt   *string         `json:"started_at,omitempty"`
	CompletedAt *string         `json:"completed_at,omitempty"`
}

type TopicDiscoveryRunView struct {
	ManuscriptID    uint64                   `json:"manuscript_id"`
	RunVersion      int                      `json:"run_version"`
	RunStatus       string                   `json:"run_status"`
	HumanCheckpoint bool                     `json:"human_checkpoint"`
	PauseAfterStage string                   `json:"pause_after_stage,omitempty"`
	Steps           []TopicDiscoveryStepView `json:"steps"`
}
