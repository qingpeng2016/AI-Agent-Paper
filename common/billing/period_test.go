package billing

import (
	"testing"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/constants"
)

func TestCountUpgradeBillingCycles(t *testing.T) {
	loc := constants.AppLocation
	d := func(y int, m time.Month, day int) time.Time {
		return time.Date(y, m, day, 0, 0, 0, 0, loc)
	}
	const days = 30

	tests := []struct {
		name               string
		periodEnd, expiresAt time.Time
		want               int
	}{
		{"current only", d(2026, 4, 26), d(2026, 4, 26), 1},
		{"before period end", d(2026, 4, 26), d(2026, 4, 10), 1},
		{"two extra cycles", d(2026, 4, 1), d(2026, 5, 31), 3},
		{"one extra cycle", d(2026, 4, 26), d(2026, 5, 26), 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountUpgradeBillingCycles(tt.periodEnd, tt.expiresAt, days)
			if got != tt.want {
				t.Fatalf("got %d want %d", got, tt.want)
			}
		})
	}
}
