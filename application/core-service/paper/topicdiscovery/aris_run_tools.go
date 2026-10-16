package topicdiscovery

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func resolveARISRepoRoot() string {
	if p := strings.TrimSpace(os.Getenv("PAPER_ARIS_REPO")); p != "" {
		return p
	}
	for _, c := range []string{"libs/aris", filepath.Join("AI-Agent-Paper", "libs/aris")} {
		if st, err := os.Stat(filepath.Join(c, "tools", "run_state.py")); err == nil && !st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return ""
}

func arisPythonBin() string {
	if p := strings.TrimSpace(os.Getenv("PAPER_ARIS_PYTHON")); p != "" {
		return p
	}
	return "python3"
}

func runARISTool(paperRoot, scriptRel string, args ...string) (stdout, stderr string, err error) {
	repo := resolveARISRepoRoot()
	if repo == "" {
		return "", "", fmt.Errorf("未找到 ARIS tools（设置 PAPER_ARIS_REPO 或确保 libs/aris/tools 存在）")
	}
	script := filepath.Join(repo, scriptRel)
	if st, err := os.Stat(script); err != nil || st.IsDir() {
		return "", "", fmt.Errorf("缺少 ARIS 脚本: %s", scriptRel)
	}
	cmdArgs := append([]string{script}, args...)
	cmd := exec.Command(arisPythonBin(), cmdArgs...)
	cmd.Dir = paperRoot
	cmd.Env = append(os.Environ(), "PYTHONPATH="+repo)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return outBuf.String(), errBuf.String(), fmt.Errorf("%s: %w (%s)", scriptRel, err, strings.TrimSpace(errBuf.String()))
	}
	return outBuf.String(), errBuf.String(), nil
}

func ensureIdeaDiscoveryRunStateStarted(paperRoot, runID string, cfg ARISIdeaDiscoveryConfig) error {
	statePath := filepath.Join(paperRoot, ".aris", "runs", runID+".json")
	if st, err := os.Stat(statePath); err == nil && st.Size() > 0 {
		return nil
	}
	executor := strings.TrimSpace(os.Getenv("PAPER_ARIS_CLI_MODEL"))
	if executor == "" {
		executor = strings.TrimSpace(os.Getenv("PAPER_ARIS_EXECUTOR"))
	}
	if executor == "" {
		executor = "claude"
	}
	phases := cfg.runStatePhasesCSV()
	args := []string{
		"start", ".", runID,
		"--phases", phases,
		"--executor", executor,
	}
	if cfg.RunState.ProvisionalAdvances {
		args = append(args, "--provisional-advances")
	}
	_, _, err := runARISTool(paperRoot, "tools/run_state.py", args...)
	return err
}

// ideaDiscoveryResumePhase 返回 "" 表示全部 phase 已 terminal（可跑 gate）；否则为待续 phase 名。
func ideaDiscoveryResumePhase(paperRoot, runID string) (string, error) {
	out, stderr, err := runARISTool(paperRoot, "tools/run_state.py", "resume", ".", runID)
	if err != nil {
		return "", err
	}
	line := strings.TrimSpace(strings.Split(out, "\n")[0])
	if line == "COMPLETE" {
		return "", nil
	}
	if line == "" {
		return "", fmt.Errorf("run_state resume 无输出: %s", stderr)
	}
	return line, nil
}

type ideaDiscoveryGateOutcome struct {
	Passed  bool
	Reasons []string
	Stdout  string
}

func runIdeaDiscoveryEvidenceGate(paperRoot, runID, reportRel string) (ideaDiscoveryGateOutcome, error) {
	reportRel = strings.TrimSpace(reportRel)
	if reportRel == "" {
		reportRel = "idea-stage/IDEA_REPORT.md"
	}
	out, stderr, err := runARISTool(
		paperRoot,
		"tools/idea_discovery_gate.py",
		".", runID,
		"--report", reportRel,
	)
	combined := strings.TrimSpace(out + "\n" + stderr)
	if err != nil {
		reasons := []string{err.Error()}
		for _, ln := range strings.Split(combined, "\n") {
			ln = strings.TrimSpace(ln)
			if strings.HasPrefix(ln, "BLOCKED:") {
				reasons = append(reasons, strings.TrimPrefix(ln, "BLOCKED:"))
			}
		}
		return ideaDiscoveryGateOutcome{Passed: false, Reasons: reasons, Stdout: combined}, err
	}
	passed := strings.Contains(combined, "PASS: idea-discovery evidence gate")
	return ideaDiscoveryGateOutcome{Passed: passed, Stdout: combined}, nil
}
