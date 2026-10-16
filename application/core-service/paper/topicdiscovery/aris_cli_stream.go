package topicdiscovery

import (
	"bufio"
	"context"
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

// streamARISCLIOutput 按行写入 info.log（stdout/warn 级 stderr）或 error.log（stderr），并经由 zap 打到控制台。
func streamARISCLIOutput(ctx context.Context, r io.Reader, stream string, acc *strings.Builder) {
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

func runARISCLIWithStream(ctx context.Context, cmd *exec.Cmd) (stdout, stderr string, err error) {
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
		streamARISCLIOutput(ctx, stdoutPipe, "stdout", &outBuf)
	}()
	go func() {
		defer wg.Done()
		streamARISCLIOutput(ctx, stderrPipe, "stderr", &errBuf)
	}()
	if err := cmd.Start(); err != nil {
		return "", "", err
	}
	wg.Wait()
	runErr := cmd.Wait()
	return outBuf.String(), errBuf.String(), runErr
}
