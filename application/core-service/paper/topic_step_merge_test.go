package paper

import "testing"

func TestMergeStepExtraPreservesPlatformRequests(t *testing.T) {
	base := map[string]any{
		ExtraKeyLiteraturePlatformRequests: []any{
			map[string]any{"source_code": "arxiv", "error": "timeout"},
		},
		"hit_count": float64(0),
	}
	patch := map[string]any{
		"hit_count": float64(5),
		ExtraKeyLiteraturePlatformRequests: []LiteraturePlatformRequestLog{
			{SourceCode: "openalex", Input: map[string]any{"limit": 5}},
		},
	}
	out := MergeStepExtra(base, patch)
	if out["hit_count"] != float64(5) {
		t.Fatalf("hit_count: got %v", out["hit_count"])
	}
	arr, ok := out[ExtraKeyLiteraturePlatformRequests].([]any)
	if !ok || len(arr) != 2 {
		t.Fatalf("platform requests: got %v", out[ExtraKeyLiteraturePlatformRequests])
	}
}
