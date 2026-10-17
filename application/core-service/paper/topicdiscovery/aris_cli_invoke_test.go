package topicdiscovery

import (
	"strings"
	"testing"
)

func TestBuildIdeaDiscoveryCLIPrompt(t *testing.T) {
	args := `"dir" — venue: NeurIPS — AUTO_PROCEED: true — RESUMABLE: true — sources: web — ARXIV_DOWNLOAD: false — language: zh — resume run-abc`
	got := buildIdeaDiscoveryCLIPrompt(args)
	want := "/idea-discovery " + args
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
