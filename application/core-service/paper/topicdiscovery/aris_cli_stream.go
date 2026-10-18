package topicdiscovery

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"go.uber.org/zap"
)

const (
	arisCLILogLineMaxBytes = 8192
	arisCLIHeartbeatEvery  = 30 * time.Second
)

func arisCLILogEnabled() bool {
	v := strings.TrimSpace(os.Getenv("PAPER_ARIS_CLI_LOG"))
	return v == "" || v == "1" || strings.EqualFold(v, "true")
}

// arisCLIConsoleMirrorEnabled 将 ARIS 子进程 stdout/stderr 同步打到当前进程终端（与 json 输出格式无关）。
// 默认：stderr 为 TTY 时开启（本机 go run --run_bot=true）；PAPER_ARIS_CLI_CONSOLE=1|0 强制开/关。
func arisCLIConsoleMirrorEnabled() bool {
	v := strings.TrimSpace(os.Getenv("PAPER_ARIS_CLI_CONSOLE"))
	switch {
	case v == "0" || strings.EqualFold(v, "false"):
		return false
	case v == "1" || strings.EqualFold(v, "true"):
		return true
	default:
		return fileIsTerminal(os.Stderr) || fileIsTerminal(os.Stdout)
	}
}

func fileIsTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func arisCLIStreamMirrorWriter(stream string) io.Writer {
	if !arisCLIConsoleMirrorEnabled() {
		return nil
	}
	if stream == "stderr" {
		return os.Stderr
	}
	return os.Stdout
}

func truncateARISCLILogLine(line string) (logged string, fullLen int, truncated bool) {
	fullLen = len(line)
	if fullLen <= arisCLILogLineMaxBytes {
		return line, fullLen, false
	}
	return line[:arisCLILogLineMaxBytes] + "…[truncated]", fullLen, true
}

func writeARISCLILogLine(ctx context.Context, stream, line string, file *os.File) {
	line = strings.TrimRight(line, "\r")
	if line == "" {
		return
	}
	if file != nil {
		_, _ = fmt.Fprintf(file, "[%s] %s\n", stream, line)
	}
	if arisCLILogEnabled() {
		logARISCLILine(ctx, stream, line)
	}
}

func flushARISCLILogLines(ctx context.Context, stream string, pending *strings.Builder, file *os.File) {
	for {
		s := pending.String()
		idx := strings.IndexByte(s, '\n')
		if idx < 0 {
			return
		}
		line := s[:idx]
		rest := s[idx+1:]
		pending.Reset()
		pending.WriteString(rest)
		writeARISCLILogLine(ctx, stream, line, file)
	}
}

// streamARISCLIOutput 读取子进程流：按块即时 mirror 到终端；按行写入 workspace log 与 zap（超长行截断）。
func streamARISCLIOutput(ctx context.Context, r io.Reader, stream string, acc *strings.Builder, file *os.File) {
	if acc == nil {
		discard := strings.Builder{}
		acc = &discard
	}
	mirror := arisCLIStreamMirrorWriter(stream)
	var pending strings.Builder
	buf := make([]byte, 32*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			acc.Write(chunk)
			if mirror != nil {
				_, _ = mirror.Write(chunk)
			}
			pending.Write(chunk)
			flushARISCLILogLines(ctx, stream, &pending, file)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
	}
	if pending.Len() > 0 {
		writeARISCLILogLine(ctx, stream, pending.String(), file)
	}
}

func logARISCLILine(ctx context.Context, stream, line string) {
	if strings.TrimSpace(line) == "" {
		return
	}
	logged, fullLen, truncated := truncateARISCLILogLine(line)
	fields := []zap.Field{
		zap.String("stream", stream),
		zap.String("line", logged),
		zap.Int("line_len", fullLen),
	}
	if truncated {
		fields = append(fields, zap.Bool("line_truncated", true))
	}
	switch stream {
	case "stderr":
		logger.WarnZ(ctx, "aris-cli", fields...)
	default:
		logger.InfoZ(ctx, "aris-cli", fields...)
	}
}

func arisCLIHeartbeatEnabled() bool {
	v := strings.TrimSpace(os.Getenv("PAPER_ARIS_CLI_HEARTBEAT"))
	switch {
	case v == "0" || strings.EqualFold(v, "false"):
		return false
	case v == "1" || strings.EqualFold(v, "true"):
		return true
	default:
		return arisCLIConsoleMirrorEnabled()
	}
}

func arisCLIWorkspaceArtifactHint(workspace string) string {
	workspace = strings.TrimSpace(workspace)
	if workspace == "" {
		return ""
	}
	names := []string{"RESEARCH_LIT_REPORT.md", "REFERENCE_INDEX.md", "EXECUTIVE_SUMMARY.md", "PROJECT_SUMMARY.txt", "RESEARCH_BRIEF.md"}
	var parts []string
	for _, name := range names {
		p := filepath.Join(workspace, name)
		fi, err := os.Stat(p)
		if err != nil || fi.IsDir() {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s(%dB)", name, fi.Size()))
	}
	if len(parts) == 0 {
		return ""
	}
	return " artifacts: " + strings.Join(parts, ", ")
}

// arisCLIRunHeartbeat 在长耗时 json 模式下仍向终端提示进程存活（ARIS 常在结束时才输出整段 JSON）。
func arisCLIRunHeartbeat(ctx context.Context, workspace string) {
	ticker := time.NewTicker(arisCLIHeartbeatEvery)
	defer ticker.Stop()
	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(start).Round(time.Second)
			hint := arisCLIWorkspaceArtifactHint(workspace)
			_, _ = fmt.Fprintf(os.Stderr,
				"[aris] still running (%s) — json mode may stay quiet until finish; watch paper/ files%s\n",
				elapsed, hint)
		}
	}
}

func runARISCLIWithStream(ctx context.Context, cmd *exec.Cmd, logFile *os.File, workspace string) (stdout, stderr string, err error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", "", err
	}
	hbCtx, hbCancel := context.WithCancel(ctx)
	defer hbCancel()
	if arisCLIHeartbeatEnabled() {
		go arisCLIRunHeartbeat(hbCtx, workspace)
	}
	var outBuf, errBuf strings.Builder
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		streamARISCLIOutput(ctx, stdoutPipe, "stdout", &outBuf, logFile)
	}()
	go func() {
		defer wg.Done()
		streamARISCLIOutput(ctx, stderrPipe, "stderr", &errBuf, logFile)
	}()
	if err := cmd.Start(); err != nil {
		return "", "", err
	}
	wg.Wait()
	runErr := cmd.Wait()
	if logFile != nil && runErr != nil {
		_, _ = fmt.Fprintf(logFile, "--- exit ---\nerror: %v\n", runErr)
	} else if logFile != nil {
		_, _ = fmt.Fprintln(logFile, "--- exit ---\nok")
	}
	return outBuf.String(), errBuf.String(), runErr
}
