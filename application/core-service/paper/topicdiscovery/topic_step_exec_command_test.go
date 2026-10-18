package topicdiscovery

import (
	"encoding/json"
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

func TestFormatARISInvokeShellCommand_ideaDiscovery(t *testing.T) {
	cfg := defaultARISIdeaDiscoveryConfig()
	in := topicRunInput{
		Description: "研究内容：MDD 脑网络 + 关键词：graph theory",
		Keywords:    []string{"major depressive disorder"},
	}
	argsLine := cfg.FormatIdeaDiscoveryArgsLine(in, "NeurIPS", "zh")
	cliPrompt := buildIdeaDiscoveryCLIPrompt(argsLine)
	const ws = "/tmp/ws"
	line := FormatARISInvokeShellCommand(ws, cliPrompt)
	if !strings.HasPrefix(line, "cd '/tmp/ws' && aris --output-format json prompt '") {
		t.Fatalf("prefix: %q", line)
	}
	for _, sub := range []string{
		"/idea-discovery",
		"— venue: NeurIPS —",
		"AUTO_PROCEED: false",
		"RESUMABLE: true",
		"— sources: web —",
		"ARXIV_DOWNLOAD: true",
		"ARXIV_MAX_DOWNLOAD: 5",
		"— language: zh",
	} {
		if !strings.Contains(line, sub) {
			t.Fatalf("missing %q in %q", sub, line)
		}
	}
	stored := formatARISCLIExecCommandStored(ws, "aris", buildARISNativeCLIArgs(cliPrompt))
	if stored != line {
		t.Fatalf("stored mismatch:\n%q\n%q", stored, line)
	}
}

func TestAppendTopicStepExecCommand_jsonNumbered(t *testing.T) {
	step := &entity.PaperOutputTopicStep{StageCode: "idea_discovery"}
	line1 := formatARISCLIExecCommandStored("/tmp/ws", "aris", []string{"--output-format", "json", "prompt", "one"})
	line2 := formatARISCLIExecCommandStored("/tmp/ws", "aris", []string{"--output-format", "json", "prompt", "two"})
	appendTopicStepExecCommandFromRecord(step, line1)
	appendTopicStepExecCommandFromRecord(step, line2)
	if step.ExecCommand == nil {
		t.Fatal("exec_command nil")
	}
	var root map[string]any
	if err := json.Unmarshal([]byte(*step.ExecCommand), &root); err != nil {
		t.Fatalf("not json: %v raw=%q", err, *step.ExecCommand)
	}
	stage, ok := root["idea_discovery"].(map[string]any)
	if !ok {
		t.Fatalf("idea_discovery missing: %#v", root)
	}
	if stage["1"] != line1 || stage["2"] != line2 {
		t.Fatalf("expected string commands: %#v", stage)
	}
}
