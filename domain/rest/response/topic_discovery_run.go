package response

import "encoding/json"

type TopicDiscoveryStepView struct {
	StageCode        string          `json:"stage_code"`
	Sort             int             `json:"sort,omitempty"`
	Status           string          `json:"status"`
	ExecProgressPct uint8  `json:"exec_progress_pct,omitempty"`
	ErrorMsg        string `json:"error_msg,omitempty"`
	ExecCommand json.RawMessage `json:"exec_command,omitempty"`
	InputParams json.RawMessage `json:"input_params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Files  json.RawMessage `json:"files,omitempty"`
	StartedAt   *string         `json:"started_at,omitempty"`
	CompletedAt *string         `json:"completed_at,omitempty"`
}

type IdeaDiscoveryUIGroupView struct {
	Index    int      `json:"index"`
	PhaseIDs []string `json:"phase_ids"`
	LabelZh  string   `json:"label_zh"`
}

type TopicDiscoveryRunView struct {
	ManuscriptID            uint64                     `json:"manuscript_id"`
	RunVersion              int                        `json:"run_version"`
	RunStatus               string                     `json:"run_status"`
	HumanCheckpoint         bool                       `json:"human_checkpoint"`
	PauseAfterStage         string                     `json:"pause_after_stage,omitempty"`
	IdeaDiscoveryUIGroups   []IdeaDiscoveryUIGroupView `json:"idea_discovery_ui_groups,omitempty"`
	Steps                   []TopicDiscoveryStepView   `json:"steps"`
}
