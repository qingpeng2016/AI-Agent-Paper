package topicdiscovery

import (
	"context"
	"fmt"
	"strings"
)

func runIdeaDiscoveryEvidenceGate(paperRoot, runID, reportRel string) (ideaDiscoveryGateOutcome, error) {
	reportRel = strings.TrimSpace(reportRel)
	if reportRel == "" {
		reportRel = "idea-stage/IDEA_REPORT.md"
	}
	runID = strings.TrimSpace(runID)
	prompt := fmt.Sprintf(`/idea-discovery

Platform-only evidence gate (installed ARIS skill; do not start new research phases).
run_id=%s
report=%s

Run the idea-discovery evidence gate per skill contract, then reply with ONLY one JSON object:
{"gate_passed": true|false, "blocked_reasons": ["..."]}
`, runID, reportRel)

	ctx := context.Background()
	stdout, stderr, err := runARISCLI(ctx, paperRoot, prompt)
	combined := strings.TrimSpace(stdout + "\n" + stderr)
	out := ideaDiscoveryGateOutcome{Stdout: combined}
	if err != nil {
		out.Passed = false
		out.Reasons = []string{err.Error()}
		return out, err
	}
	if obj, _ := extractJSONObject(stdout); obj != nil {
		if v, ok := obj["gate_passed"].(bool); ok {
			out.Passed = v
		}
		if arr, ok := obj["blocked_reasons"].([]any); ok {
			for _, x := range arr {
				if s := strings.TrimSpace(fmt.Sprint(x)); s != "" {
					out.Reasons = append(out.Reasons, s)
				}
			}
		}
		if out.Passed {
			return out, nil
		}
		if len(out.Reasons) == 0 {
			out.Reasons = []string{"gate_passed=false"}
		}
		return out, fmt.Errorf("idea_discovery_gate 未 PASS")
	}
	out.Passed = strings.Contains(combined, "PASS: idea-discovery evidence gate")
	if !out.Passed {
		return out, fmt.Errorf("idea_discovery_gate 未 PASS")
	}
	return out, nil
}

type ideaDiscoveryGateOutcome struct {
	Passed  bool
	Reasons []string
	Stdout  string
}
