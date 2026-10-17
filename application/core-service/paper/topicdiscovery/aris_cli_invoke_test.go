package topicdiscovery

import (
	"strings"
	"testing"
)

func TestBuildIdeaDiscoveryCLIPrompt(t *testing.T) {
	args := `MDD rs-fMRI — venue: NeurIPS, AUTO_PROCEED: true, sources: web — resume run-abc`
	got := buildIdeaDiscoveryCLIPrompt(args, "")
	want := "/idea-discovery " + args
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	withNotes := buildIdeaDiscoveryCLIPrompt(args, "Platform user checkpoint reply: PROCEED")
	if !strings.Contains(withNotes, "/idea-discovery") || !strings.Contains(withNotes, args) || !strings.Contains(withNotes, "PROCEED") {
		t.Fatalf("missing parts: %q", withNotes)
	}
}
