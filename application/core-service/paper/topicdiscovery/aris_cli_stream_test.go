package topicdiscovery

import (
	"os"
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
