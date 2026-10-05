package billing

import "time"

const DefaultPeriodDays = 30

// PeriodDays 商品计费周期天数；period_days<=0 时按 billing_period 兜底。
func PeriodDays(periodDays int, billingPeriod string) int {
	if periodDays > 0 {
		return periodDays
	}
	switch billingPeriod {
	case "year":
		return 365
	case "once":
		return 30
	default:
		return DefaultPeriodDays
	}
}

// AddPeriod 从 from 起推进一个计费周期（按天）。
func AddPeriod(from time.Time, periodDays int) time.Time {
	d := periodDays
	if d <= 0 {
		d = DefaultPeriodDays
	}
	return from.AddDate(0, 0, d)
}

// AddPeriods 推进 n 个计费周期。
func AddPeriods(from time.Time, periodDays, count int) time.Time {
	t := from
	for i := 0; i < count; i++ {
		t = AddPeriod(t, periodDays)
	}
	return t
}

// CountUpgradeBillingCycles 升档计价周期数：当前周期计 1；expires_at 晚于 period_end 时自 period_end 按 period_days 累加。
func CountUpgradeBillingCycles(periodEnd, expiresAt time.Time, periodDays int) int {
	const maxQty = 99
	qty := 1
	if !expiresAt.After(periodEnd) {
		return qty
	}
	cursor := periodEnd
	for cursor.Before(expiresAt) && qty < maxQty {
		cursor = AddPeriod(cursor, periodDays)
		qty++
	}
	return qty
}
