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

const ideaDiscoveryMaxInvokeRetries = 32

var ideaDiscoveryRunStatePhases = []string{
	"research-lit",
	"idea-creator",
	"novelty-check",
	"research-review",
	"research-refine-pipeline",
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
	PauseAfterRetrieve bool   `yaml:"pause_after_retrieve"`
	AutoProceed        bool   `yaml:"auto_proceed"`
	Sources            string `yaml:"sources"`
	MaxArisPerBot      int    `yaml:"max_aris_per_bot"`
	RequireRetrieve    bool   `yaml:"require_retrieve"`
}

// ARISIdeaDiscoveryConfig 与 inout/aris_idea_discovery.yaml 对齐。
type ARISIdeaDiscoveryConfig struct {
	Artifacts ideaDiscoveryArtifacts `yaml:"artifacts"`
	Platform  ideaDiscoveryPlatform  `yaml:"platform"`
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

func defaultARISIdeaDiscoveryConfig() ARISIdeaDiscoveryConfig {
	return ARISIdeaDiscoveryConfig{
		Artifacts: defaultIdeaDiscoveryArtifacts(),
		Platform: ideaDiscoveryPlatform{
			PauseAfterRetrieve: true,
			AutoProceed:        false,
			Sources:            "local",
			MaxArisPerBot:      1,
			RequireRetrieve:    true,
		},
	}
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
	if strings.TrimSpace(c.Platform.Sources) == "" {
		c.Platform.Sources = "local"
	}
	if c.Platform.MaxArisPerBot <= 0 {
		c.Platform.MaxArisPerBot = 1
	}
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

func (c ARISIdeaDiscoveryConfig) formatSkillArgumentsLine(paperRoot, dir, venue string) string {
	ap := "false"
	if c.Platform.AutoProceed {
		ap = "true"
	}
	composed := strings.TrimSpace(c.Artifacts.IdeaReport)
	line := fmt.Sprintf(
		`%s — venue: %s, AUTO_PROCEED: %s, RESUMABLE: true, ARXIV_DOWNLOAD: false, render html: true, compact: false, composed: %s`,
		dir, venue, ap, composed,
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
	return line
}
