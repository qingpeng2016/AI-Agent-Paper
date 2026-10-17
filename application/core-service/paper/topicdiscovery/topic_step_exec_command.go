package topicdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
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

func buildARISCLIExecEntry(paperRoot, cli string, args []string, prompt string) map[string]any {
	shellLine := formatARISCLIExecShellLine(cli, args)
	entry := map[string]any{
		"invoked_at":    time.Now().UTC().Format(time.RFC3339),
		"workspace":     filepath.Clean(paperRoot),
		"command":       shellLine,
		"exec_command":  shellLine,
	}
	if p := strings.TrimSpace(prompt); p != "" {
		entry["prompt"] = p
		if strings.HasPrefix(p, "/idea-discovery") {
			entry["behavior"] = "idea-discovery"
		}
	}
	return entry
}

func formatARISCLIExecRecord(paperRoot, cli string, args []string, prompt string) string {
	entry := buildARISCLIExecEntry(paperRoot, cli, args, prompt)
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "=== ARIS CLI invoke %s ===\n", entry["invoked_at"])
	_, _ = fmt.Fprintf(&b, "workspace: %s\n", entry["workspace"])
	_, _ = fmt.Fprintf(&b, "command: %s\n", entry["command"])
	if p, _ := entry["prompt"].(string); p != "" {
		_, _ = fmt.Fprintln(&b, "--- prompt ---")
		_, _ = fmt.Fprintln(&b, p)
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func parseARISExecRecordText(record string) map[string]any {
	record = strings.TrimSpace(record)
	if record == "" {
		return nil
	}
	entry := map[string]any{}
	for _, line := range strings.Split(record, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "===") || line == "--- prompt ---" {
			continue
		}
		if strings.HasPrefix(line, "workspace:") {
			entry["workspace"] = strings.TrimSpace(strings.TrimPrefix(line, "workspace:"))
			continue
		}
		if strings.HasPrefix(line, "command:") {
			cmd := strings.TrimSpace(strings.TrimPrefix(line, "command:"))
			entry["command"] = cmd
			entry["exec_command"] = cmd
			continue
		}
	}
	if idx := strings.Index(record, "--- prompt ---"); idx >= 0 {
		prompt := strings.TrimSpace(record[idx+len("--- prompt ---"):])
		if prompt != "" {
			entry["prompt"] = prompt
		}
	}
	if len(entry) == 0 {
		entry["raw"] = record
	}
	return entry
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
		// 兼容 json.Unmarshal 到 map[string]any 时 nested 为 map[string]interface{}
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

func appendTopicStepExecCommandEntry(step *entity.PaperOutputTopicStep, entry map[string]any) {
	if step == nil || len(entry) == 0 {
		return
	}
	stage := strings.TrimSpace(step.StageCode)
	if stage == "" {
		stage = "unknown"
	}
	root := loadExecCommandRoot(step)
	stageMap := stageExecCommandMap(root, stage)
	stageMap[nextExecCommandIndex(stageMap)] = entry
	root[stage] = stageMap
	saveExecCommandRoot(step, root)
}

func appendTopicStepExecCommandFromRecord(step *entity.PaperOutputTopicStep, record string) {
	entry := parseARISExecRecordText(record)
	if len(entry) == 0 {
		return
	}
	appendTopicStepExecCommandEntry(step, entry)
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
