package topicdiscovery

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestArisCLIConsoleMirrorEnabled_envOverride(t *testing.T) {
	t.Setenv("PAPER_ARIS_CLI_CONSOLE", "1")
	if !arisCLIConsoleMirrorEnabled() {
		t.Fatal("expected force on")
	}
	t.Setenv("PAPER_ARIS_CLI_CONSOLE", "0")
	if arisCLIConsoleMirrorEnabled() {
		t.Fatal("expected force off")
	}
}

func TestFileIsTerminal_stdErr(t *testing.T) {
	_ = fileIsTerminal(os.Stderr)
}

func TestTruncateARISCLILogLine(t *testing.T) {
	short := "hello"
	logged, n, truncated := truncateARISCLILogLine(short)
	if logged != short || n != len(short) || truncated {
		t.Fatalf("unexpected %q %d %v", logged, n, truncated)
	}
	long := strings.Repeat("x", arisCLILogLineMaxBytes+100)
	logged, n, truncated = truncateARISCLILogLine(long)
	if !truncated || n != len(long) || len(logged) <= arisCLILogLineMaxBytes {
		t.Fatalf("expected truncation, got len=%d truncated=%v", len(logged), truncated)
	}
}

func TestStreamARISCLIOutput_chunksAccumulate(t *testing.T) {
	t.Setenv("PAPER_ARIS_CLI_CONSOLE", "0")
	t.Setenv("PAPER_ARIS_CLI_LOG", "0")
	var acc strings.Builder
	r := strings.NewReader("line1\npartial")
	streamARISCLIOutput(context.Background(), r, "stdout", &acc, nil)
	if acc.String() != "line1\npartial" {
		t.Fatalf("acc=%q", acc.String())
	}
}
