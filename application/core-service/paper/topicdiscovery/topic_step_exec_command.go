package topicdiscovery

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

const maxTopicStepExecCommandBytes = 512 * 1024

func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// formatARISCLIExecShellLine 可复制的 shell 单行（prompt 为完整参数）。
func formatARISCLIExecShellLine(cli string, args []string) string {
	cli = strings.TrimSpace(cli)
	if cli == "" {
		cli = "aris"
	}
	if len(args) == 0 {
		return cli
	}
	parts := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.EqualFold(a, "prompt") && i+1 < len(args) {
			parts = append(parts, "prompt", shellSingleQuote(args[i+1]))
			i++
			continue
		}
		if a == "-p" && i+1 < len(args) {
			parts = append(parts, "-p", shellSingleQuote(args[i+1]))
			i++
			continue
		}
		if strings.ContainsAny(a, " \t\n\"'$`\\*?[]") {
			parts = append(parts, shellSingleQuote(a))
		} else {
			parts = append(parts, a)
		}
	}
	return cli + " " + strings.Join(parts, " ")
}

func formatARISCLIExecRecord(paperRoot, cli string, args []string, prompt string) string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "=== ARIS CLI invoke %s ===\n", time.Now().UTC().Format(time.RFC3339))
	_, _ = fmt.Fprintf(&b, "workspace: %s\n", filepath.Clean(paperRoot))
	_, _ = fmt.Fprintf(&b, "command: %s\n", formatARISCLIExecShellLine(cli, args))
	prompt = strings.TrimSpace(prompt)
	if prompt != "" {
		_, _ = fmt.Fprintln(&b, "--- prompt ---")
		_, _ = fmt.Fprintln(&b, prompt)
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func appendTopicStepExecCommand(step *entity.PaperOutputTopicStep, record string) {
	record = strings.TrimSpace(record)
	if step == nil || record == "" {
		return
	}
	block := record
	if !strings.HasSuffix(block, "\n") {
		block += "\n"
	}
	var prev string
	if step.ExecCommand != nil {
		prev = *step.ExecCommand
	}
	combined := prev + block
	combined = truncateBytes(combined, maxTopicStepExecCommandBytes)
	step.ExecCommand = &combined
}

func (s *TopicDiscoveryRunService) persistTopicStepExecCommand(ctx context.Context, step *entity.PaperOutputTopicStep, record string) {
	if s == nil || step == nil || strings.TrimSpace(record) == "" {
		return
	}
	appendTopicStepExecCommand(step, record)
	if s.steps != nil {
		_ = s.steps.SaveStep(ctx, step)
	}
}
