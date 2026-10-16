package topicdiscovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/storage"
	"gopkg.in/yaml.v3"
)

// ARISIdeaDiscoveryConfig 与 inout/aris_idea_discovery.yaml 对齐原生 SKILL 契约。
type ARISIdeaDiscoveryConfig struct {
	Artifacts struct {
		IdeaReport        string `yaml:"idea_report"`
		FinalProposal     string `yaml:"final_proposal"`
		ExperimentPlan    string `yaml:"experiment_plan"`
		ExperimentTracker string `yaml:"experiment_tracker"`
		PipelineSummary   string `yaml:"pipeline_summary"`
		ResearchBrief     string `yaml:"research_brief"`
		LiteratureDir     string `yaml:"literature_dir"`
		CorpusJSON        string `yaml:"corpus_json"`
	} `yaml:"artifacts"`
	RunState struct {
		Phases              []string `yaml:"phases"`
		ProvisionalAdvances bool     `yaml:"provisional_advances"`
	} `yaml:"run_state"`
	Skill struct {
		AutoProceed       bool   `yaml:"auto_proceed"`
		Resumable         bool   `yaml:"resumable"`
		ArxivDownload     bool   `yaml:"arxiv_download"`
		RenderHTML        bool   `yaml:"render_html"`
		Compact           bool   `yaml:"compact"`
		Sources           string `yaml:"sources"`
		ComposedReport    string `yaml:"composed_report"`
		ReviewerModel     string `yaml:"reviewer_model"`
		OutputDir         string `yaml:"output_dir"`
		PilotMaxHours     int    `yaml:"pilot_max_hours"`
		PilotTimeoutHours int    `yaml:"pilot_timeout_hours"`
		MaxPilotIdeas     int    `yaml:"max_pilot_ideas"`
		MaxTotalGPUHours  int    `yaml:"max_total_gpu_hours"`
	} `yaml:"skill"`
	Platform struct {
		RequireRetrieveBeforeIdeaDiscovery bool `yaml:"require_retrieve_before_idea_discovery"`
		InjectStagedLiterature             bool `yaml:"inject_staged_literature"`
		AutoChainRetrieve                  bool `yaml:"auto_chain_retrieve"`
		AutoPersistAfterGate               bool `yaml:"auto_persist_after_gate"`
		InterStepHumanCheckpoint           bool `yaml:"inter_step_human_checkpoint"`
	} `yaml:"platform"`
	Invoke struct {
		TimeoutSec               int  `yaml:"timeout_sec"`
		MaxInvokeRetries         int  `yaml:"max_invoke_retries"`
		MaxInvokesPerRun         int  `yaml:"max_invokes_per_run"`
		RequireFinalPlatformJSON bool `yaml:"require_final_platform_json"`
		NativeHeadless           bool `yaml:"native_headless"`
	} `yaml:"invoke"`
	Gate struct {
		Report string `yaml:"report"`
	} `yaml:"gate"`
}

var (
	ideaDiscoveryCfg     ARISIdeaDiscoveryConfig
	ideaDiscoveryCfgOnce sync.Once
	ideaDiscoveryCfgErr  error
)

func defaultARISIdeaDiscoveryConfig() ARISIdeaDiscoveryConfig {
	var c ARISIdeaDiscoveryConfig
	c.Artifacts.IdeaReport = "idea-stage/IDEA_REPORT.md"
	c.Artifacts.FinalProposal = "refine-logs/FINAL_PROPOSAL.md"
	c.Artifacts.ExperimentPlan = "refine-logs/EXPERIMENT_PLAN.md"
	c.Artifacts.ExperimentTracker = "refine-logs/EXPERIMENT_TRACKER.md"
	c.Artifacts.PipelineSummary = "refine-logs/PIPELINE_SUMMARY.md"
	c.Artifacts.ResearchBrief = "RESEARCH_BRIEF.md"
	c.Artifacts.LiteratureDir = "literature"
	c.Artifacts.CorpusJSON = ".paper/corpus.json"
	c.RunState.Phases = []string{
		"research-lit", "idea-creator", "novelty-check", "research-review", "research-refine-pipeline",
	}
	c.RunState.ProvisionalAdvances = false
	c.Skill.AutoProceed = false
	c.Skill.Resumable = true
	c.Skill.ArxivDownload = false
	c.Platform.RequireRetrieveBeforeIdeaDiscovery = true
	c.Platform.InjectStagedLiterature = false
	c.Platform.AutoChainRetrieve = false
	c.Platform.AutoPersistAfterGate = false
	c.Platform.InterStepHumanCheckpoint = true
	c.Skill.RenderHTML = true
	c.Skill.Compact = false
	c.Skill.Sources = "local"
	c.Skill.ComposedReport = "idea-stage/IDEA_REPORT.md"
	c.Skill.ReviewerModel = "gpt-6-astra"
	c.Skill.OutputDir = "idea-stage/"
	c.Skill.PilotMaxHours = 2
	c.Skill.PilotTimeoutHours = 3
	c.Skill.MaxPilotIdeas = 3
	c.Skill.MaxTotalGPUHours = 8
	c.Invoke.TimeoutSec = 0
	c.Invoke.MaxInvokeRetries = 32
	c.Invoke.MaxInvokesPerRun = 1
	c.Invoke.RequireFinalPlatformJSON = false
	c.Invoke.NativeHeadless = true
	c.Gate.Report = "idea-stage/IDEA_REPORT.md"
	return c
}

func ideaDiscoveryConfigPath() string {
	if p := strings.TrimSpace(os.Getenv("PAPER_ARIS_IDEA_DISCOVERY_CONFIG")); p != "" {
		return p
	}
	return filepath.Join(storage.PaperInoutRoot(), "aris_idea_discovery.yaml")
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
	if len(c.RunState.Phases) == 0 {
		c.RunState.Phases = defaultARISIdeaDiscoveryConfig().RunState.Phases
	}
	if strings.TrimSpace(c.Gate.Report) == "" {
		c.Gate.Report = c.Artifacts.IdeaReport
	}
	if strings.TrimSpace(c.Skill.ComposedReport) == "" {
		c.Skill.ComposedReport = c.Artifacts.IdeaReport
	}
	// timeout_sec: 0 = 不限制（原生长 session）；可用 PAPER_ARIS_INVOKE_TIMEOUT_SEC 覆盖
	if c.Invoke.MaxInvokeRetries <= 0 {
		c.Invoke.MaxInvokeRetries = 12
	}
	if c.Invoke.MaxInvokesPerRun <= 0 {
		c.Invoke.MaxInvokesPerRun = c.Invoke.MaxInvokeRetries
		if c.Invoke.MaxInvokesPerRun <= 0 {
			c.Invoke.MaxInvokesPerRun = 32
		}
	}
}

func (c ARISIdeaDiscoveryConfig) runStatePhasesCSV() string {
	parts := make([]string, 0, len(c.RunState.Phases))
	for _, p := range c.RunState.Phases {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, ",")
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

func (c ARISIdeaDiscoveryConfig) formatSkillArgumentsLine(paperRoot, dir, venue string) string {
	ap := "false"
	if c.Skill.AutoProceed {
		ap = "true"
	}
	res := "false"
	if c.Skill.Resumable {
		res = "true"
	}
	arxiv := "false"
	if c.Skill.ArxivDownload {
		arxiv = "true"
	}
	rhtml := "false"
	if c.Skill.RenderHTML {
		rhtml = "true"
	}
	compact := "false"
	if c.Skill.Compact {
		compact = "true"
	}
	composed := strings.TrimSpace(c.Skill.ComposedReport)
	line := fmt.Sprintf(
		`%s — venue: %s, AUTO_PROCEED: %s, RESUMABLE: %s, ARXIV_DOWNLOAD: %s, render html: %s, compact: %s, composed: %s`,
		dir, venue, ap, res, arxiv, rhtml, compact, composed,
	)
	if src := strings.TrimSpace(c.Skill.Sources); src != "" {
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
	return line
}
