package topicdiscovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateIdeaDiscoveryPlatformEvidence 平台确定性门禁：不依赖 LLM，只检查 skill 约定的路径与报告锚节。
func ValidateIdeaDiscoveryPlatformEvidence(paperRoot string, cfg ARISIdeaDiscoveryConfig) (passed bool, reasons []string) {
	paperRoot = strings.TrimSpace(paperRoot)
	if paperRoot == "" {
		return false, []string{"paperRoot 为空"}
	}
	ev := cfg.normalizedEvidenceGate()
	for rel, min := range ev.MinFileBytes {
		rel = strings.TrimSpace(rel)
		if rel == "" {
			continue
		}
		if !workspaceFileMinSize(paperRoot, rel, min) {
			reasons = append(reasons, fmt.Sprintf("缺少或过小: %s (需要 ≥ %d bytes)", rel, min))
		}
	}
	reportRel := strings.TrimSpace(cfg.Artifacts.IdeaReport)
	if reportRel == "" {
		reportRel = "idea-stage/IDEA_REPORT.md"
	}
	reportPath := filepath.Join(paperRoot, reportRel)
	reportBody, _ := os.ReadFile(reportPath)
	if len(reportBody) == 0 && len(ev.ReportHeadings) > 0 {
		reasons = append(reasons, fmt.Sprintf("无法读取报告以校验章节: %s", reportRel))
	} else {
		text := string(reportBody)
		for _, h := range ev.ReportHeadings {
			h = strings.TrimSpace(h)
			if h == "" {
				continue
			}
			if !ideaReportContainsHeading(text, h) {
				reasons = append(reasons, fmt.Sprintf("IDEA_REPORT 缺少章节: %s", h))
			}
		}
	}
	return len(reasons) == 0, reasons
}

func ideaReportContainsHeading(body, heading string) bool {
	body = strings.ToLower(body)
	heading = strings.ToLower(strings.TrimSpace(heading))
	if heading == "" {
		return true
	}
	// Markdown "## Foo" or HTML anchor id literature-landscape
	slug := strings.NewReplacer("#", "", " ", "-").Replace(heading)
	slug = strings.Trim(slug, "-")
	candidates := []string{heading}
	if strings.HasPrefix(heading, "##") {
		candidates = append(candidates, strings.TrimSpace(strings.TrimPrefix(heading, "##")))
	} else {
		candidates = append(candidates, "## "+heading)
	}
	if slug != "" {
		candidates = append(candidates, slug, "#"+slug)
	}
	for _, c := range candidates {
		c = strings.TrimSpace(c)
		if c != "" && strings.Contains(body, c) {
			return true
		}
	}
	return false
}

func (c ARISIdeaDiscoveryConfig) normalizedEvidenceGate() ideaDiscoveryEvidenceGate {
	def := defaultIdeaDiscoveryEvidenceGate(c)
	ev := c.EvidenceGate
	out := ideaDiscoveryEvidenceGate{
		MinFileBytes:   map[string]int64{},
		ReportHeadings: ev.ReportHeadings,
	}
	if len(ev.MinFileBytes) > 0 {
		for k, v := range ev.MinFileBytes {
			if strings.TrimSpace(k) == "" || v <= 0 {
				continue
			}
			out.MinFileBytes[k] = v
		}
	} else {
		out.MinFileBytes = def.MinFileBytes
	}
	if len(out.ReportHeadings) == 0 {
		out.ReportHeadings = def.ReportHeadings
	}
	return out
}

func defaultIdeaDiscoveryEvidenceGate(cfg ARISIdeaDiscoveryConfig) ideaDiscoveryEvidenceGate {
	a := cfg.Artifacts
	if strings.TrimSpace(a.IdeaReport) == "" {
		a = defaultIdeaDiscoveryArtifacts()
	}
	return ideaDiscoveryEvidenceGate{
		MinFileBytes: map[string]int64{
			a.IdeaReport:        2048,
			a.FinalProposal:     512,
			a.ExperimentPlan:    512,
			a.ExperimentTracker: 256,
		},
		ReportHeadings: []string{
			"Literature Landscape",
			"Ranked Ideas",
			"Novelty Verification",
			"External Critical Review",
			"Executive Summary",
		},
	}
}
