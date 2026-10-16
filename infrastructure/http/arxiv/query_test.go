package arxiv

import (
	"strings"
	"testing"
)

func TestBuildArxivSearchQueryParam_ANDNotOR(t *testing.T) {
	got := buildArxivSearchQueryParam("major depressive disorder resting-state fMRI")
	if strings.Contains(got, "all:major depressive") {
		t.Fatal("must not pass whole phrase as single all:")
	}
	for _, sub := range []string{" AND ", "all:major", "all:depressive", "all:resting-state"} {
		if !strings.Contains(got, sub) {
			t.Fatalf("missing %q in %s", sub, got)
		}
	}
}
