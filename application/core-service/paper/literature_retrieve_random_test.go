package paper

import "testing"

func TestPickUnusedRandomSlot_offset(t *testing.T) {
	used := map[int]struct{}{0: {}}
	slot := pickUnusedRandomSlot(100, 2, literaturePagingOffset, used)
	if slot == 0 {
		t.Fatalf("expected unused slot != 0, got 0")
	}
	if _, ok := used[slot]; !ok {
		t.Fatalf("slot %d not marked used", slot)
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
