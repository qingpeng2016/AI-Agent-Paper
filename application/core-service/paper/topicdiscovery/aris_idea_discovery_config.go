package topicdiscovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/lang"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/storage"
	"gopkg.in/yaml.v3"
)

const ideaDiscoveryMaxInvokeRetries = 32

var ideaDiscoveryRunStatePhases = []string{
	"research-lit",
	"idea-creator",
	"novelty-check",
	"research-review",
	"research-refine-pipeline",
}

// IdeaDiscoveryCheckpointPhase 与 conf/aris_idea_discovery.yaml 顶层 checkpoint_phases 每项对齐。
type IdeaDiscoveryCheckpointPhase struct {
	Pause   bool   `yaml:"pause"`
	LabelZh string `yaml:"label_zh"`
}

func defaultIdeaDiscoveryCheckpointPhases() map[string]IdeaDiscoveryCheckpointPhase {
	return map[string]IdeaDiscoveryCheckpointPhase{
		"research-lit":             {Pause: false, LabelZh: "文献调研"},
		"idea-creator":             {Pause: false, LabelZh: "选题大纲"},
		"novelty-check":            {Pause: true, LabelZh: "新颖性检验"},
		"research-review":          {Pause: true, LabelZh: "外部审查"},
		"research-refine-pipeline": {Pause: true, LabelZh: "文献综述与实验规划"},
	}
}

func normalizeIdeaDiscoveryCheckpointPhases(m map[string]IdeaDiscoveryCheckpointPhase) map[string]IdeaDiscoveryCheckpointPhase {
	def := defaultIdeaDiscoveryCheckpointPhases()
	out := make(map[string]IdeaDiscoveryCheckpointPhase, len(ideaDiscoveryRunStatePhases))
	for _, phase := range ideaDiscoveryRunStatePhases {
		d := def[phase]
		if m != nil {
			if v, ok := m[phase]; ok {
				entry := IdeaDiscoveryCheckpointPhase{
					Pause:   v.Pause,
					LabelZh: strings.TrimSpace(v.LabelZh),
				}
				if entry.LabelZh == "" {
					entry.LabelZh = d.LabelZh
				}
				out[phase] = entry
				continue
			}
		}
		out[phase] = d
	}
	return out
}

type ideaDiscoveryArtifacts struct {
	IdeaReport        string `yaml:"idea_report"`
	FinalProposal     string `yaml:"final_proposal"`
	ExperimentPlan    string `yaml:"experiment_plan"`
	ExperimentTracker string `yaml:"experiment_tracker"`
	PipelineSummary   string `yaml:"pipeline_summary"`
	ResearchBrief     string `yaml:"research_brief"`
	LiteratureDir     string `yaml:"literature_dir"`
	CorpusJSON        string `yaml:"corpus_json"`
}

type ideaDiscoveryPlatform struct {
	AutoProceed   bool   `yaml:"auto_proceed"`
	// Resume：run_state 已存在时是否在 platform_user_prompt 追加「— resume <run_id>」（段间续跑仍靠 run_state + session notes）。
	Resume        bool   `yaml:"resume"`
	Sources       string `yaml:"sources"`
	ArxivDownload bool   `yaml:"arxiv_download"`
	MaxArisPerBot int    `yaml:"max_aris_per_bot"`
}

type ideaDiscoveryWorkspace struct {
	InvokeDir     string `yaml:"invoke_dir"`
	InvokeLogsDir string `yaml:"invoke_logs_dir"`
	RunStateDir   string `yaml:"run_state_dir"`
}

// arisRunWorkspaceDir inout/.../run_N/ 下 ARIS 工作区目录（与后续 skill 共用，产物以子路径区分）。
const arisRunWorkspaceDir = "paper"

// ARISIdeaDiscoveryConfig 与 conf/aris_idea_discovery.yaml 对齐。
type ARISIdeaDiscoveryConfig struct {
	Artifacts          ideaDiscoveryArtifacts                  `yaml:"artifacts"`
	Workspace          ideaDiscoveryWorkspace                  `yaml:"workspace"`
	CheckpointPhases   map[string]IdeaDiscoveryCheckpointPhase `yaml:"checkpoint_phases"`
	CLISessionCleanup  IdeaDiscoveryCLISessionCleanup          `yaml:"cli_session_cleanup"`
	Platform           ideaDiscoveryPlatform                   `yaml:"platform"`
}

var (
	ideaDiscoveryCfg     ARISIdeaDiscoveryConfig
	ideaDiscoveryCfgOnce sync.Once
	ideaDiscoveryCfgErr  error
)

func defaultIdeaDiscoveryArtifacts() ideaDiscoveryArtifacts {
	return ideaDiscoveryArtifacts{
		IdeaReport:        "idea-stage/IDEA_REPORT.md",
		FinalProposal:     "refine-logs/FINAL_PROPOSAL.md",
		ExperimentPlan:    "refine-logs/EXPERIMENT_PLAN.md",
		ExperimentTracker: "refine-logs/EXPERIMENT_TRACKER.md",
		PipelineSummary:   "refine-logs/PIPELINE_SUMMARY.md",
		ResearchBrief:     "RESEARCH_BRIEF.md",
		LiteratureDir:     "literature",
		CorpusJSON:        ".paper/corpus.json",
	}
}

func defaultIdeaDiscoveryWorkspace() ideaDiscoveryWorkspace {
	return ideaDiscoveryWorkspace{
		InvokeDir:     ".paper/invoke",
		InvokeLogsDir: ".paper/invoke/logs",
		RunStateDir:   ".aris/runs",
	}
}

func paperRootJoin(paperRoot, rel string) string {
	paperRoot = filepath.Clean(paperRoot)
	rel = strings.TrimSpace(filepath.ToSlash(rel))
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" {
		return paperRoot
	}
	return filepath.Join(paperRoot, filepath.FromSlash(rel))
}

// InvokeDirAbs paperRoot 下 invoke 目录（绝对路径）。
func (c ARISIdeaDiscoveryConfig) InvokeDirAbs(paperRoot string) string {
	rel := strings.TrimSpace(c.Workspace.InvokeDir)
	if rel == "" {
		rel = defaultIdeaDiscoveryWorkspace().InvokeDir
	}
	return paperRootJoin(paperRoot, rel)
}

// InvokeLogsDirAbs CLI 日志目录（绝对路径）。
func (c ARISIdeaDiscoveryConfig) InvokeLogsDirAbs(paperRoot string) string {
	rel := strings.TrimSpace(c.Workspace.InvokeLogsDir)
	if rel == "" {
		rel = defaultIdeaDiscoveryWorkspace().InvokeLogsDir
	}
	return paperRootJoin(paperRoot, rel)
}

// RunStateFileAbs run_state JSON 文件（绝对路径）。
func (c ARISIdeaDiscoveryConfig) RunStateFileAbs(paperRoot, runID string) string {
	rel := strings.TrimSpace(c.Workspace.RunStateDir)
	if rel == "" {
		rel = defaultIdeaDiscoveryWorkspace().RunStateDir
	}
	runID = strings.TrimSpace(runID)
	return filepath.Join(paperRootJoin(paperRoot, rel), runID+".json")
}

// ShouldAppendCLIResume 是否在 $ARGUMENTS 追加「— resume <run_id>」。
func (c ARISIdeaDiscoveryConfig) ShouldAppendCLIResume(paperRoot, runID string) bool {
	if !c.Platform.Resume {
		return false
	}
	return ideaDiscoveryShouldAppendResume(paperRoot, runID)
}

func (c ARISIdeaDiscoveryConfig) invokeDirRel() string {
	rel := strings.TrimSpace(c.Workspace.InvokeDir)
	if rel == "" {
		return defaultIdeaDiscoveryWorkspace().InvokeDir
	}
	return filepath.ToSlash(rel)
}

func defaultARISIdeaDiscoveryConfig() ARISIdeaDiscoveryConfig {
	return ARISIdeaDiscoveryConfig{
		Artifacts: defaultIdeaDiscoveryArtifacts(),
		Workspace: defaultIdeaDiscoveryWorkspace(),
		CheckpointPhases:  defaultIdeaDiscoveryCheckpointPhases(),
		CLISessionCleanup: defaultCLISessionCleanup(),
		Platform: ideaDiscoveryPlatform{
			AutoProceed:   false,
			Resume:        false,
			Sources:       "web",
			ArxivDownload: true,
			MaxArisPerBot: 1,
		},
	}
}

// PhaseNeedsUserCheckpoint 下一待执行 phase 是否触发平台人工检查点（checkpoint_phases[phase].pause）。
func (c ARISIdeaDiscoveryConfig) PhaseNeedsUserCheckpoint(pendingPhase string) bool {
	pendingPhase = strings.TrimSpace(pendingPhase)
	if pendingPhase == "" {
		return false
	}
	m := c.CheckpointPhases
	if len(m) == 0 {
		m = defaultIdeaDiscoveryCheckpointPhases()
	}
	if entry, ok := m[pendingPhase]; ok {
		return entry.Pause
	}
	return true
}

// PhaseLabelZh 返回 phase 展示中文名（conf checkpoint_phases.label_zh）。
func (c ARISIdeaDiscoveryConfig) PhaseLabelZh(phase string) string {
	phase = strings.TrimSpace(phase)
	if phase == "" {
		return ""
	}
	m := c.CheckpointPhases
	if len(m) == 0 {
		m = defaultIdeaDiscoveryCheckpointPhases()
	}
	if entry, ok := m[phase]; ok {
		if s := strings.TrimSpace(entry.LabelZh); s != "" {
			return s
		}
	}
	return phase
}

// IdeaDiscoveryCheckpointPhaseList 按 run_state 顺序返回 phase 配置（供 API/前端展示）。
func (c ARISIdeaDiscoveryConfig) IdeaDiscoveryCheckpointPhaseList() []struct {
	ID      string
	Pause   bool
	LabelZh string
} {
	m := normalizeIdeaDiscoveryCheckpointPhases(c.CheckpointPhases)
	out := make([]struct {
		ID      string
		Pause   bool
		LabelZh string
	}, 0, len(ideaDiscoveryRunStatePhases))
	for _, id := range ideaDiscoveryRunStatePhases {
		e := m[id]
		out = append(out, struct {
			ID      string
			Pause   bool
			LabelZh string
		}{ID: id, Pause: e.Pause, LabelZh: e.LabelZh})
	}
	return out
}

func ideaDiscoveryConfigPath() string {
	if p := strings.TrimSpace(os.Getenv("PAPER_ARIS_IDEA_DISCOVERY_CONFIG")); p != "" {
		return p
	}
	return filepath.Join(storage.PaperConfRoot(), "aris_idea_discovery.yaml")
}

func LoadARISIdeaDiscoveryConfig() (ARISIdeaDiscoveryConfig, error) {
	ideaDiscoveryCfgOnce.Do(func() {
		ideaDiscoveryCfg = defaultARISIdeaDiscoveryConfig()
		path := ideaDiscoveryConfigPath()
		b, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				ideaDiscoveryCfgErr = nil
				normalizeIdeaDiscoveryConfig(&ideaDiscoveryCfg)
				return
			}
			ideaDiscoveryCfgErr = fmt.Errorf("read %s: %w", path, err)
			return
		}
		if err := yaml.Unmarshal(b, &ideaDiscoveryCfg); err != nil {
			ideaDiscoveryCfgErr = fmt.Errorf("parse %s: %w", path, err)
			return
		}
		normalizeIdeaDiscoveryConfig(&ideaDiscoveryCfg)
	})
	return ideaDiscoveryCfg, ideaDiscoveryCfgErr
}

func normalizeIdeaDiscoveryConfig(c *ARISIdeaDiscoveryConfig) {
	def := defaultIdeaDiscoveryArtifacts()
	a := &c.Artifacts
	if strings.TrimSpace(a.IdeaReport) == "" {
		a.IdeaReport = def.IdeaReport
	}
	if strings.TrimSpace(a.FinalProposal) == "" {
		a.FinalProposal = def.FinalProposal
	}
	if strings.TrimSpace(a.ExperimentPlan) == "" {
		a.ExperimentPlan = def.ExperimentPlan
	}
	if strings.TrimSpace(a.ExperimentTracker) == "" {
		a.ExperimentTracker = def.ExperimentTracker
	}
	if strings.TrimSpace(a.PipelineSummary) == "" {
		a.PipelineSummary = def.PipelineSummary
	}
	if strings.TrimSpace(a.ResearchBrief) == "" {
		a.ResearchBrief = def.ResearchBrief
	}
	if strings.TrimSpace(a.LiteratureDir) == "" {
		a.LiteratureDir = def.LiteratureDir
	}
	if strings.TrimSpace(a.CorpusJSON) == "" {
		a.CorpusJSON = def.CorpusJSON
	}
	wdef := defaultIdeaDiscoveryWorkspace()
	w := &c.Workspace
	if strings.TrimSpace(w.InvokeDir) == "" {
		w.InvokeDir = wdef.InvokeDir
	}
	if strings.TrimSpace(w.InvokeLogsDir) == "" {
		w.InvokeLogsDir = wdef.InvokeLogsDir
	}
	if strings.TrimSpace(w.RunStateDir) == "" {
		w.RunStateDir = wdef.RunStateDir
	}
	if strings.TrimSpace(c.Platform.Sources) == "" {
		c.Platform.Sources = "local"
	}
	if c.Platform.MaxArisPerBot <= 0 {
		c.Platform.MaxArisPerBot = 1
	}
	c.CheckpointPhases = normalizeIdeaDiscoveryCheckpointPhases(c.CheckpointPhases)
	normalizeCLISessionCleanup(&c.CLISessionCleanup)
}

func ideaDiscoveryRunStatePhasesCSV() string {
	return strings.Join(ideaDiscoveryRunStatePhases, ",")
}

func (c ARISIdeaDiscoveryConfig) artifactPathsForResult() []string {
	seen := map[string]struct{}{}
	var out []string
	for _, p := range []string{
		c.Artifacts.IdeaReport,
		c.Artifacts.FinalProposal,
		c.Artifacts.ExperimentPlan,
		c.Artifacts.ExperimentTracker,
		c.Artifacts.PipelineSummary,
	} {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func (c ARISIdeaDiscoveryConfig) artifactConfigMap() map[string]string {
	a := c.Artifacts
	return map[string]string{
		"idea_report":        a.IdeaReport,
		"final_proposal":     a.FinalProposal,
		"experiment_plan":    a.ExperimentPlan,
		"experiment_tracker": a.ExperimentTracker,
		"pipeline_summary":   a.PipelineSummary,
		"research_brief":     a.ResearchBrief,
		"literature_dir":     a.LiteratureDir,
		"corpus_json":        a.CorpusJSON,
		"gate_report":        a.IdeaReport,
		"composed_report":    a.IdeaReport,
		"run_state_phases":   ideaDiscoveryRunStatePhasesCSV(),
	}
}

func (c ARISIdeaDiscoveryConfig) formatSkillArgumentsLine(paperRoot, dir, venue, language string) string {
	ap := "false"
	if c.Platform.AutoProceed {
		ap = "true"
	}
	composed := strings.TrimSpace(c.Artifacts.IdeaReport)
	arxivDL := "false"
	if c.Platform.ArxivDownload {
		arxivDL = "true"
	}
	line := fmt.Sprintf(
		`%s — venue: %s, AUTO_PROCEED: %s, RESUMABLE: true, ARXIV_DOWNLOAD: %s, render html: true, compact: false, composed: %s`,
		dir, venue, ap, arxivDL, composed,
	)
	if src := strings.TrimSpace(c.Platform.Sources); src != "" {
		line += fmt.Sprintf(`, sources: %s`, src)
	}
	litDir := strings.TrimSpace(c.Artifacts.LiteratureDir)
	if litDir == "" {
		litDir = "literature"
	}
	if strings.TrimSpace(paperRoot) != "" {
		lib := filepath.Join(filepath.Clean(paperRoot), filepath.FromSlash(litDir))
		if abs, err := filepath.Abs(lib); err == nil {
			lib = abs
		}
		line += fmt.Sprintf(`, paper library: %s`, lib)
	}
	if loc := lang.NormalizeContentLanguage(language); loc != "" {
		line += fmt.Sprintf(`, language: %s`, loc)
	}
	return line
}
