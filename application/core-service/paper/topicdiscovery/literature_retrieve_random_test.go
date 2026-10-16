package topicdiscovery

import "testing"

func TestPickUnusedRandomSlot_offset(t *testing.T) {
	used := map[int]struct{}{0: {}}
	slot, pageIdx, pageCount := pickUnusedRandomSlot(100, literaturePagingOffset, sourceCrossref, used)
	if pageCount != 10 {
		t.Fatalf("pageCount %d want 10", pageCount)
	}
	if slot == 0 || slot%literatureRandomPageSize != 0 {
		t.Fatalf("slot %d should be non-zero multiple of %d", slot, literatureRandomPageSize)
	}
	if pageIdx != slot/literatureRandomPageSize {
		t.Fatalf("pageIdx %d slot %d mismatch", pageIdx, slot)
	}
}

func TestLiteratureSequentialAdvanceSlot(t *testing.T) {
	if got := literatureSequentialAdvanceSlot(sourceArxiv, 0); got != 10 {
		t.Fatalf("arxiv slot %d want 10", got)
	}
	if got := literatureSequentialAdvanceSlot(sourceOpenAlex, 1); got != 2 {
		t.Fatalf("openalex page %d want 2", got)
	}
}

func TestLiteratureRandomRoundLimits_smallBatch(t *testing.T) {
	lim := literatureRandomRoundLimits(3)
	want := [5]int{1, 1, 1, 0, 0}
	if lim != want {
		t.Fatalf("3 papers random: got %v want %v", lim, want)
	}
}

func TestPickUnusedRandomSlot_offsetCap(t *testing.T) {
	for i := 0; i < 30; i++ {
		slot, _, pageCount := pickUnusedRandomSlot(1_456_560, literaturePagingOffset, sourceCrossref, map[int]struct{}{})
		if slot > literatureRandomMaxOffset {
			t.Fatalf("slot %d exceeds cap %d", slot, literatureRandomMaxOffset)
		}
		if slot%literatureRandomPageSize != 0 {
			t.Fatalf("slot %d not aligned to page size", slot)
		}
		maxPages := literatureRandomMaxOffset/literatureRandomPageSize + 1
		if pageCount > maxPages {
			t.Fatalf("pageCount %d exceeds cap %d", pageCount, maxPages)
		}
	}
}

func TestLiteratureRandomPageCount(t *testing.T) {
	if n := literatureRandomPageCount(64, literaturePagingOffset, sourceCrossref); n != 7 {
		t.Fatalf("64 items: got %d pages want 7", n)
	}
	if n := literatureRandomPageCount(1_456_560, literaturePagingOffset, sourceCrossref); n != literatureRandomMaxOffset/literatureRandomPageSize+1 {
		t.Fatalf("huge total: page cap got %d", n)
	}
	if n := literatureRandomPageCount(1_456_560, literaturePagingOffset, sourceArxiv); n != literatureRandomArxivMaxPages {
		t.Fatalf("arxiv huge total: got %d pages want %d", n, literatureRandomArxivMaxPages)
	}
}

func TestPickUnusedRandomSlot_arxivShallow(t *testing.T) {
	for i := 0; i < 40; i++ {
		slot, _, pageCount := pickUnusedRandomSlot(1_459_981, literaturePagingOffset, sourceArxiv, map[int]struct{}{})
		if pageCount > literatureRandomArxivMaxPages {
			t.Fatalf("pageCount %d exceeds arxiv cap", pageCount)
		}
		if slot >= literatureRandomArxivMaxPages*literatureRandomPageSize {
			t.Fatalf("slot %d too deep for arxiv random", slot)
		}
	}
}

func TestLiteratureRoundLimits(t *testing.T) {
	lim := literatureRoundLimits(10)
	want := [5]int{2, 2, 2, 2, 2}
	if lim != want {
		t.Fatalf("10 papers: got %v want %v", lim, want)
	}
	lim = literatureRoundLimits(11)
	want = [5]int{3, 2, 2, 2, 2}
	if lim != want {
		t.Fatalf("11 papers: got %v want %v", lim, want)
	}
	lim = literatureRoundLimits(3)
	want = [5]int{3, 0, 0, 0, 0}
	if lim != want {
		t.Fatalf("3 papers: got %v want %v", lim, want)
	}
}
