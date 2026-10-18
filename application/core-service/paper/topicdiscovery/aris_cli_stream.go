package topicdiscovery

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"go.uber.org/zap"
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

func mirrorARISCLILine(stream, line string) {
	if !arisCLIConsoleMirrorEnabled() {
		return
	}
	out := os.Stdout
	if stream == "stderr" {
		out = os.Stderr
	}
	_, _ = fmt.Fprintln(out, line)
}

// streamARISCLIOutput 按行写入 info.log（stdout/warn 级 stderr）或 error.log（stderr），并经由 zap 打到控制台。
func streamARISCLIOutput(ctx context.Context, r io.Reader, stream string, acc *strings.Builder, file *os.File) {
	if acc == nil {
		discard := strings.Builder{}
		acc = &discard
	}
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		acc.WriteString(line)
		acc.WriteByte('\n')
		if file != nil {
			_, _ = fmt.Fprintf(file, "[%s] %s\n", stream, line)
		}
		mirrorARISCLILine(stream, line)
		if arisCLILogEnabled() {
			logARISCLILine(ctx, stream, line)
		}
	}
}

func logARISCLILine(ctx context.Context, stream, line string) {
	if strings.TrimSpace(line) == "" {
		return
	}
	fields := []zap.Field{
		zap.String("stream", stream),
		zap.String("line", line),
	}
	switch stream {
	case "stderr":
		logger.WarnZ(ctx, "aris-cli", fields...)
	default:
		logger.InfoZ(ctx, "aris-cli", fields...)
	}
}

func runARISCLIWithStream(ctx context.Context, cmd *exec.Cmd, logFile *os.File) (stdout, stderr string, err error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", "", err
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
