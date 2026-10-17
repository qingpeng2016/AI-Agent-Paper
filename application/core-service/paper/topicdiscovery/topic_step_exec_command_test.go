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
	argsLine := cfg.FormatIdeaDiscoveryArgsLine(in, "NeurIPS", "zh", "", false)
	cliPrompt := buildIdeaDiscoveryCLIPrompt(argsLine)
	line := FormatARISInvokeShellCommand(cliPrompt)
	if !strings.HasPrefix(line, "aris --output-format json prompt '") {
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
	entry := buildARISCLIExecEntry("/tmp/ws", "aris", buildARISNativeCLIArgs(cliPrompt), cliPrompt)
	if entry["exec_command"] != entry["command"] {
		t.Fatalf("exec_command mismatch: %#v", entry)
	}
}

func TestAppendTopicStepExecCommand_jsonNumbered(t *testing.T) {
	step := &entity.PaperOutputTopicStep{StageCode: "idea_discovery"}
	rec1 := formatARISCLIExecRecord("/tmp/ws", "aris", []string{"prompt", "one"}, "one")
	rec2 := formatARISCLIExecRecord("/tmp/ws", "aris", []string{"prompt", "two"}, "two")
	appendTopicStepExecCommandFromRecord(step, rec1)
	appendTopicStepExecCommandFromRecord(step, rec2)
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
	if stage["1"] == nil || stage["2"] == nil {
		t.Fatalf("expected keys 1 and 2: %#v", stage)
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
