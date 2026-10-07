package arxiv

import "testing"

func TestEntryToPaper(t *testing.T) {
	p := entryToPaper(entryXML{
		ID:        "http://arxiv.org/abs/2401.12345v1",
		Title:     "  Sparse\n  Attention ",
		Summary:   "An abstract.",
		Published: "2024-01-15T12:00:00Z",
		Authors:   []authorXML{{Name: "A. Author"}},
	})
	if p.ExternalKey != "arxiv:2401.12345v1" {
		t.Fatalf("external key %q", p.ExternalKey)
	}
	if p.PublishedYear != 2024 {
		t.Fatalf("year %d", p.PublishedYear)
	}
}
