package topicdiscovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func arisCLIWorkspaceLogEnabled() bool {
	v := strings.TrimSpace(os.Getenv("PAPER_ARIS_CLI_WORKSPACE_LOG"))
	if v == "0" || strings.EqualFold(v, "false") {
		return false
	}
	return true
}

func arisInvokeLogsDir(paperRoot string) string {
	return mustIdeaDiscoveryConfig().InvokeLogsDirAbs(paperRoot)
}

func newARISCLILogFile(paperRoot, cli string) (*os.File, string, error) {
	dir := arisInvokeLogsDir(paperRoot)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, "", err
	}
	stamp := time.Now().UTC().Format("20060102T150405.000Z")
	safeCLI := sanitizePathSegment(filepath.Base(cli))
	if safeCLI == "" || safeCLI == "unknown" {
		safeCLI = "cli"
	}
	name := fmt.Sprintf("%s_%s.log", stamp, safeCLI)
	path := filepath.Join(dir, name)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, "", err
	}
	return f, path, nil
}

func writeARISCLILogHeader(f *os.File, paperRoot, cli string, args []string, prompt string) {
	if f == nil {
		return
	}
	_, _ = fmt.Fprintf(f, "=== ARIS CLI invoke %s ===\n", time.Now().UTC().Format(time.RFC3339))
	_, _ = fmt.Fprintf(f, "workspace: %s\n", filepath.Clean(paperRoot))
	_, _ = fmt.Fprintf(f, "command: %s %s\n", cli, strings.Join(args, " "))
	prompt = strings.TrimSpace(prompt)
	if len(prompt) > 2000 {
		inv := mustIdeaDiscoveryConfig().invokeDirRel()
		prompt = prompt[:2000] + fmt.Sprintf("\n…(prompt truncated; see %s/platform_user_prompt.txt)", inv)
	}
	if prompt != "" {
		_, _ = fmt.Fprintf(f, "--- prompt ---\n%s\n", prompt)
	}
	_, _ = fmt.Fprintln(f, "--- stream ---")
}

// appendARISInvokeErrorLog 追加 invoke 失败摘要（与 .paper/invoke/logs/*_aris_cli.log 互补）。
func appendARISInvokeErrorLog(paperRoot string, invokeErr error, retry int) {
	if !arisCLIWorkspaceLogEnabled() || invokeErr == nil {
		return
	}
	dir := arisInvokeLogsDir(paperRoot)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	path := filepath.Join(dir, "invoke_errors.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "=== %s retry=%d ===\n%s\n\n", time.Now().UTC().Format(time.RFC3339), retry, invokeErr.Error())
	_ = os.WriteFile(filepath.Join(dir, "latest_invoke_error.txt"), []byte(invokeErr.Error()), 0o644)
}

func finalizeARISCLILog(paperRoot, logPath string) {
	if !arisCLIWorkspaceLogEnabled() || strings.TrimSpace(logPath) == "" {
		return
	}
	b, err := os.ReadFile(logPath)
	if err != nil {
		return
	}
	latest := filepath.Join(arisInvokeLogsDir(paperRoot), "latest_aris_cli.log")
	_ = os.WriteFile(latest, b, 0o644)
}
