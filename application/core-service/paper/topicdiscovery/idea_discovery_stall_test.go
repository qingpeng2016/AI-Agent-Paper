package topicdiscovery

import (
	"strings"
	"testing"
)

func TestIdeaDiscoveryStallBumpAndClear(t *testing.T) {
	meta := map[string]any{}
	if got := ideaDiscoveryStallCount(meta, "research-lit"); got != 0 {
		t.Fatalf("count: %d", got)
	}
	if n := ideaDiscoveryStallBump(meta, "research-lit"); n != 1 {
		t.Fatalf("bump: %d", n)
	}
	if n := ideaDiscoveryStallBump(meta, "research-lit"); n != 2 {
		t.Fatalf("bump2: %d", n)
	}
	ideaDiscoveryStallClear(meta, "research-lit")
	if got := ideaDiscoveryStallCount(meta, "research-lit"); got != 0 {
		t.Fatalf("after clear: %d", got)
	}
}

func TestIsNonRetriableIdeaDiscoveryInvokeError_promptTooLong(t *testing.T) {
	msg := "aris failed: error: anthropic api returned 400 Bad Request (invalid_request_error): prompt is too long: 206864 tokens > 200000 maximum"
	if !isNonRetriableIdeaDiscoveryInvokeError(strings.ToLower(msg)) {
		t.Fatal("prompt too long should be non-retriable")
	}
}
