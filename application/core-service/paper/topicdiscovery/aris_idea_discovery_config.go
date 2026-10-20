package topicdiscovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/lang"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/storage"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"go.uber.org/zap"
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
		"research-lit":             {Pause: true, LabelZh: "文献调研"},
		"idea-creator":             {Pause: true, LabelZh: "选题大纲"},
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
	AutoProceed        bool   `yaml:"auto_proceed"`
	Sources            string `yaml:"sources"`
	ArxivDownload      bool   `yaml:"arxiv_download"`
	ArxivMaxDownload   int    `yaml:"arxiv_max_download"`
	MaxArisPerBot      int    `yaml:"max_aris_per_bot"`
	PilotMaxHours      int    `yaml:"pilot_max_hours"`
	PilotTimeoutHours  int    `yaml:"pilot_timeout_hours"`
	MaxPilotIdeas      int    `yaml:"max_pilot_ideas"`
	MaxTotalGPUHours   int    `yaml:"max_total_gpu_hours"`
	ReviewerModel      string `yaml:"reviewer_model"`
	OutputDir          string `yaml:"output_dir"`
	Compact            bool   `yaml:"compact"`
	RenderHTML         bool   `yaml:"render_html"`
	RefPaper           string `yaml:"ref_paper"`
	Resumable          bool   `yaml:"resumable"`
}

// ideaDiscoveryEvidenceGate 平台侧产物门禁（与 ARIS skill RESUMABLE gate 互补）。
type ideaDiscoveryEvidenceGate struct {
	MinFileBytes   map[string]int64 `yaml:"min_file_bytes"`
	ReportHeadings []string         `yaml:"report_headings"`
}

type ideaDiscoveryWorkspace struct {
	// InoutRoot ARIS 会话根（user_*/manuscript_*/run_*）；须在 Git 仓库外，见 conf/aris_idea_discovery.yaml。
	InoutRoot     string `yaml:"inout_root"`
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
	EvidenceGate       ideaDiscoveryEvidenceGate               `yaml:"evidence_gate"`
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

func (c ARISIdeaDiscoveryConfig) invokeDirRel() string {
	rel := strings.TrimSpace(c.Workspace.InvokeDir)
	if rel == "" {
		return defaultIdeaDiscoveryWorkspace().InvokeDir
	}
	return filepath.ToSlash(rel)
}

// InoutRootAbs 配置项 workspace.inout_root（绝对路径）；空则 storage.PaperInoutRoot()。
func (c ARISIdeaDiscoveryConfig) InoutRootAbs() string {
	if v := strings.TrimSpace(os.Getenv("PAPER_ARIS_WORK_ROOT")); v != "" {
		return absEnsureDir(v)
	}
	if v := strings.TrimSpace(os.Getenv("PAPER_INOUT_ROOT")); v != "" {
		return absEnsureDir(v)
	}
	rel := strings.TrimSpace(c.Workspace.InoutRoot)
	if rel == "" {
		return storage.PaperInoutRoot()
	}
	return absEnsureDir(expandConfigPath(rel))
}

func absEnsureDir(root string) string {
	root = strings.TrimSpace(root)
	if root == "" {
		return storage.PaperInoutRoot()
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	_, statErr := os.Stat(abs)
	missing := os.IsNotExist(statErr)
	if err := os.MkdirAll(abs, 0o755); err != nil {
		logger.ErrorZ(context.Background(), "idea-discovery-inout-root-mkdir",
			zap.String("path", abs),
			zap.Error(err))
	} else if missing {
		logger.InfoZ(context.Background(), "idea-discovery-inout-root-ready",
			zap.String("path", abs))
	}
	return abs
}

func defaultARISIdeaDiscoveryConfig() ARISIdeaDiscoveryConfig {
	return ARISIdeaDiscoveryConfig{
		Artifacts: defaultIdeaDiscoveryArtifacts(),
		Workspace: defaultIdeaDiscoveryWorkspace(),
		CheckpointPhases:  defaultIdeaDiscoveryCheckpointPhases(),
		CLISessionCleanup: defaultCLISessionCleanup(),
		Platform: ideaDiscoveryPlatform{
			AutoProceed:       false,
			Sources:           "web",
			ArxivDownload:     true,
			ArxivMaxDownload:  5,
			MaxArisPerBot:     1,
			PilotMaxHours:     2,
			PilotTimeoutHours: 3,
			MaxPilotIdeas:     3,
			MaxTotalGPUHours:  8,
			ReviewerModel:     "gpt-6-astra",
			OutputDir:         "idea-stage/",
			Compact:           false,
			RenderHTML:        true,
			RefPaper:          "",
			Resumable:         true,
		},
		EvidenceGate: defaultIdeaDiscoveryEvidenceGate(ARISIdeaDiscoveryConfig{Artifacts: defaultIdeaDiscoveryArtifacts()}),
	}
}

// FirstCheckpointPhaseID 与 conf checkpoint_phases 顺序一致的首个 ARIS phase（默认创建 run 即视为用户已确认该段）。
func (c ARISIdeaDiscoveryConfig) FirstCheckpointPhaseID() string {
	if len(ideaDiscoveryRunStatePhases) == 0 {
		return "research-lit"
	}
	return ideaDiscoveryRunStatePhases[0]
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
	if c.Platform.ArxivMaxDownload <= 0 {
		c.Platform.ArxivMaxDownload = defaultARISIdeaDiscoveryConfig().Platform.ArxivMaxDownload
	}
	pdef := defaultARISIdeaDiscoveryConfig().Platform
	if c.Platform.PilotMaxHours <= 0 {
		c.Platform.PilotMaxHours = pdef.PilotMaxHours
	}
	if c.Platform.PilotTimeoutHours <= 0 {
		c.Platform.PilotTimeoutHours = pdef.PilotTimeoutHours
	}
	if c.Platform.MaxPilotIdeas <= 0 {
		c.Platform.MaxPilotIdeas = pdef.MaxPilotIdeas
	}
	if c.Platform.MaxTotalGPUHours <= 0 {
		c.Platform.MaxTotalGPUHours = pdef.MaxTotalGPUHours
	}
	if strings.TrimSpace(c.Platform.ReviewerModel) == "" {
		c.Platform.ReviewerModel = pdef.ReviewerModel
	}
	if strings.TrimSpace(c.Platform.OutputDir) == "" {
		c.Platform.OutputDir = pdef.OutputDir
	}
	c.CheckpointPhases = normalizeIdeaDiscoveryCheckpointPhases(c.CheckpointPhases)
	normalizeCLISessionCleanup(&c.CLISessionCleanup)
	// 启动时创建 workspace.inout_root（含 /www/.../AI-Agent-Paper-runtime/inout），无需手工 mkdir。
	_ = c.InoutRootAbs()
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
		"run_state_phases":     ideaDiscoveryRunStatePhasesCSV(),
		"arxiv_download":       fmt.Sprintf("%t", c.Platform.ArxivDownload),
		"arxiv_max_download":   fmt.Sprintf("%d", c.Platform.ArxivMaxDownload),
	}
}

// ideaDiscoveryQuotedDirection 引号内研究方向：优先 Description（原样），否则 Direction + 关键词。
func ideaDiscoveryQuotedDirection(input topicRunInput) string {
	inner := strings.TrimSpace(input.Description)
	if inner == "" {
		body := strings.TrimSpace(input.Direction)
		kw := strings.TrimSpace(keywordsLine(input))
		if kw != "" {
			inner = fmt.Sprintf("研究内容：%s + 关键词：%s", body, kw)
		} else {
			inner = body
		}
	}
	inner = strings.ReplaceAll(inner, `"`, `\"`)
	return `"` + inner + `"`
}

// FormatIdeaDiscoveryArgsLine 与 ARIS /idea-discovery skill Constants 对齐（composed 仍由 skill 内传给子 skill，不在此重复）。
func (c ARISIdeaDiscoveryConfig) FormatIdeaDiscoveryArgsLine(input topicRunInput, venue, language string) string {
	p := c.Platform
	ap := "false"
	if p.AutoProceed {
		ap = "true"
	}
	resumable := "false"
	if p.Resumable {
		resumable = "true"
	}
	arxivDL := "false"
	if p.ArxivDownload {
		arxivDL = "true"
	}
	compact := "false"
	if p.Compact {
		compact = "true"
	}
	renderHTML := "false"
	if p.RenderHTML {
		renderHTML = "true"
	}
	src := strings.TrimSpace(p.Sources)
	if src == "" {
		src = "web"
	}
	loc := lang.NormalizeContentLanguage(language)
	if loc == "" {
		loc = "zh"
	}
	venue = strings.TrimSpace(venue)
	if venue == "" {
		venue = strings.TrimSpace(input.Venue)
	}
	if venue == "" {
		venue = "NeurIPS"
	}
	outDir := strings.TrimSpace(p.OutputDir)
	if outDir == "" {
		outDir = "idea-stage/"
	}
	reviewer := strings.TrimSpace(p.ReviewerModel)
	if reviewer == "" {
		reviewer = "gpt-6-astra"
	}
	line := fmt.Sprintf(
		`%s — venue: %s — AUTO_PROCEED: %s — RESUMABLE: %s — sources: %s — ARXIV_DOWNLOAD: %s — COMPACT: %s — RENDER_HTML: %s — OUTPUT_DIR: %s — REVIEWER_MODEL: %s — PILOT_MAX_HOURS: %d — PILOT_TIMEOUT_HOURS: %d — MAX_PILOT_IDEAS: %d — MAX_TOTAL_GPU_HOURS: %d`,
		ideaDiscoveryQuotedDirection(input),
		venue,
		ap,
		resumable,
		src,
		arxivDL,
		compact,
		renderHTML,
		outDir,
		reviewer,
		p.PilotMaxHours,
		p.PilotTimeoutHours,
		p.MaxPilotIdeas,
		p.MaxTotalGPUHours,
	)
	if p.ArxivDownload {
		maxDL := p.ArxivMaxDownload
		if maxDL <= 0 {
			maxDL = 5
		}
		line += fmt.Sprintf(` — ARXIV_MAX_DOWNLOAD: %d`, maxDL)
	}
	if ref := strings.TrimSpace(p.RefPaper); ref != "" && !strings.EqualFold(ref, "false") {
		line += fmt.Sprintf(` — REF_PAPER: %s`, ref)
	}
	line += fmt.Sprintf(` — language: %s`, loc)
	return line
}
