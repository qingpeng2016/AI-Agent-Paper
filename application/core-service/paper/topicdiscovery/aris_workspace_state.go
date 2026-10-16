package topicdiscovery

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

var errNoRunStateFile = errors.New("run state file missing")

// ideaDiscoveryResumePhase 只读工作区 .aris/runs/{run_id}.json（由 aris /idea-discovery 写入），不调用 libs/aris 脚本。
func ideaDiscoveryResumePhase(paperRoot, runID string) (string, error) {
	phase, complete, err := readRunStateResumeTarget(paperRoot, runID)
	if err != nil {
		if errors.Is(err, errNoRunStateFile) {
			if len(ideaDiscoveryRunStatePhases) == 0 {
				return "", nil
			}
			return ideaDiscoveryRunStatePhases[0], nil
		}
		return "", err
	}
	if complete {
		return "", nil
	}
	return phase, nil
}

func readRunStateResumeTarget(paperRoot, runID string) (phase string, complete bool, err error) {
	path := mustIdeaDiscoveryConfig().RunStateFileAbs(paperRoot, runID)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, errNoRunStateFile
		}
		return "", false, err
	}
	var st struct {
		Phases []struct {
			Phase  string `json:"phase"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"phases"`
	}
	if err := json.Unmarshal(b, &st); err != nil {
		return "", false, err
	}
	for _, ph := range st.Phases {
		id := strings.TrimSpace(ph.Phase)
		if id == "" {
			id = strings.TrimSpace(ph.Name)
		}
		if id == "" {
			continue
		}
		if runStatePhaseNeedsWork(ph.Status) {
			return id, false, nil
		}
	}
	return "", true, nil
}

func runStatePhaseNeedsWork(status string) bool {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "accepted", "skipped":
		return false
	case "provisional":
		return false
	default:
		return true
	}
}
