package literature

import "testing"

func TestExecProgressPct(t *testing.T) {
	if got := ExecProgressPct(0, 4, true); got != 1 {
		t.Fatalf("kickstart got %d want 1", got)
	}
	if got := ExecProgressPct(2, 4, false); got != 50 {
		t.Fatalf("half got %d want 50", got)
	}
	if got := ExecProgressPct(4, 4, false); got != 100 {
		t.Fatalf("done got %d want 100", got)
	}
}
