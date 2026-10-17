package topicdiscovery

import (
	"strings"
	"testing"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

func TestFormatARISCLIExecShellLine_promptQuoted(t *testing.T) {
	line := formatARISCLIExecShellLine("aris", []string{"--output-format", "json", "prompt", "hello world"})
	if !strings.Contains(line, "prompt 'hello world'") {
		t.Fatalf("unexpected line: %q", line)
	}
}

func TestAppendTopicStepExecCommand_cumulative(t *testing.T) {
	step := &entity.PaperOutputTopicStep{}
	appendTopicStepExecCommand(step, "first\n")
	appendTopicStepExecCommand(step, "second\n")
	if step.ExecCommand == nil || !strings.Contains(*step.ExecCommand, "first") || !strings.Contains(*step.ExecCommand, "second") {
		t.Fatalf("exec_command=%q", deref(step.ExecCommand))
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
