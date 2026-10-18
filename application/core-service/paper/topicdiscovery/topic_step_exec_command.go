package topicdiscovery

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

const maxTopicStepExecCommandBytes = 512 * 1024

func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// formatARISCLIExecShellLine 可复制的 argv 单行（prompt 为完整参数）。
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

// formatARISCLIExecCommandStored 与 runARISCLI 一致：cmd.Dir=paperRoot + exec(cli, args…)。
func formatARISCLIExecCommandStored(paperRoot, cli string, args []string) string {
	line := formatARISCLIExecShellLine(cli, args)
	root := filepath.Clean(strings.TrimSpace(paperRoot))
	if root == "" {
		return line
	}
	return "cd " + shellSingleQuote(root) + " && " + line
}

func execCommandLineFromPersistRecord(record string) string {
	record = strings.TrimSpace(record)
	if record == "" {
		return ""
	}
	if !strings.HasPrefix(record, "=== ARIS CLI invoke") {
		return record
	}
	var workspace, command string
	var promptLines []string
	inPrompt := false
	for _, line := range strings.Split(record, "\n") {
		if strings.HasPrefix(line, "=== ARIS CLI invoke") {
			continue
		}
		if strings.HasPrefix(line, "workspace:") {
			workspace = strings.TrimSpace(strings.TrimPrefix(line, "workspace:"))
			inPrompt = false
			continue
		}
		if strings.HasPrefix(line, "command:") {
			command = strings.TrimSpace(strings.TrimPrefix(line, "command:"))
			inPrompt = false
			continue
		}
		if line == "--- prompt ---" {
			inPrompt = true
			continue
		}
		if inPrompt {
			promptLines = append(promptLines, line)
		}
	}
	command = strings.TrimSpace(command)
	if command == "" {
		return ""
	}
	// 旧版 record 的 command 行可能未带 shell 引号；有 prompt 段则以 prompt 为准重建。
	if len(promptLines) > 0 {
		prompt := strings.TrimSpace(strings.Join(promptLines, "\n"))
		if prompt != "" {
			cli, args := resolveARISCLIInvocation(prompt)
			command = formatARISCLIExecShellLine(cli, args)
		}
	}
	if ws := filepath.Clean(workspace); ws != "" && !strings.HasPrefix(command, "cd ") {
		return "cd " + shellSingleQuote(ws) + " && " + command
	}
	return command
}

func loadExecCommandRoot(step *entity.PaperOutputTopicStep) map[string]any {
	root := map[string]any{}
	if step == nil || step.ExecCommand == nil {
		return root
	}
	raw := strings.TrimSpace(*step.ExecCommand)
	if raw == "" {
		return root
	}
	if json.Unmarshal([]byte(raw), &root) == nil && len(root) > 0 {
		return root
	}
	if raw != "" {
		root["_legacy_plaintext"] = raw
	}
	return root
}

func stageExecCommandMap(root map[string]any, stageCode string) map[string]any {
	if root == nil {
		root = map[string]any{}
	}
	stageCode = strings.TrimSpace(stageCode)
	if stageCode == "" {
		stageCode = "unknown"
	}
	cur, ok := root[stageCode]
	if !ok {
		m := map[string]any{}
		root[stageCode] = m
		return m
	}
	switch m := cur.(type) {
	case map[string]any:
		return m
	default:
		b, _ := json.Marshal(m)
		var out map[string]any
		if json.Unmarshal(b, &out) == nil && out != nil {
			root[stageCode] = out
			return out
		}
		fresh := map[string]any{}
		root[stageCode] = fresh
		return fresh
	}
}

func nextExecCommandIndex(stageMap map[string]any) string {
	maxN := 0
	for k := range stageMap {
		n, err := strconv.Atoi(strings.TrimSpace(k))
		if err != nil || n <= 0 {
			continue
		}
		if n > maxN {
			maxN = n
		}
	}
	return strconv.Itoa(maxN + 1)
}

func appendTopicStepExecCommandLine(step *entity.PaperOutputTopicStep, commandLine string) {
	commandLine = strings.TrimSpace(commandLine)
	if step == nil || commandLine == "" {
		return
	}
	stage := strings.TrimSpace(step.StageCode)
	if stage == "" {
		stage = "unknown"
	}
	root := loadExecCommandRoot(step)
	stageMap := stageExecCommandMap(root, stage)
	stageMap[nextExecCommandIndex(stageMap)] = commandLine
	root[stage] = stageMap
	saveExecCommandRoot(step, root)
}

func appendTopicStepExecCommandFromRecord(step *entity.PaperOutputTopicStep, record string) {
	line := execCommandLineFromPersistRecord(record)
	if line == "" {
		return
	}
	appendTopicStepExecCommandLine(step, line)
}

func saveExecCommandRoot(step *entity.PaperOutputTopicStep, root map[string]any) {
	if step == nil || root == nil {
		return
	}
	for len(root) > 0 {
		b, err := json.Marshal(root)
		if err != nil {
			return
		}
		if len(b) <= maxTopicStepExecCommandBytes {
			s := string(b)
			step.ExecCommand = &s
			return
		}
		if !trimOldestExecCommandInvoke(root) {
			s := truncateBytes(string(b), maxTopicStepExecCommandBytes)
			step.ExecCommand = &s
			return
		}
	}
}

func trimOldestExecCommandInvoke(root map[string]any) bool {
	trimmed := false
	for stage, v := range root {
		if strings.HasPrefix(stage, "_") {
			continue
		}
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		minN := 0
		minKey := ""
		for k := range m {
			n, err := strconv.Atoi(strings.TrimSpace(k))
			if err != nil || n <= 0 {
				continue
			}
			if minN == 0 || n < minN {
				minN = n
				minKey = k
			}
		}
		if minKey != "" {
			delete(m, minKey)
			trimmed = true
		}
	}
	if !trimmed {
		delete(root, "_legacy_plaintext")
		return len(root) > 0
	}
	return true
}

func (s *TopicDiscoveryRunService) persistTopicStepExecCommand(ctx context.Context, step *entity.PaperOutputTopicStep, record string) {
	if s == nil || step == nil || strings.TrimSpace(record) == "" {
		return
	}
	appendTopicStepExecCommandFromRecord(step, record)
	if s.steps != nil {
		_ = s.steps.SaveStep(ctx, step)
	}
}
