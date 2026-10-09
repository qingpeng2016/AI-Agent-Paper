package paper

import "testing"

func TestResolveLiteratureHitURL_arxiv(t *testing.T) {
	got := resolveLiteratureHitURL(sourceArxiv, "arxiv:2401.12345v1", "", "")
	if got != "https://arxiv.org/abs/2401.12345v1" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveLiteratureHitURL_doi(t *testing.T) {
	got := resolveLiteratureHitURL(sourceOpenAlex, "W123", "", "10.1038/nature123")
	if got != "https://doi.org/10.1038/nature123" {
		t.Fatalf("got %q", got)
	}
}
