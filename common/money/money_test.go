package money

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestMulQty(t *testing.T) {
	unit := decimal.RequireFromString("89.00")
	got := MulQty(unit, 3)
	want := decimal.RequireFromString("267.00")
	if !got.Equal(want) {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestAddInvoiceSurcharge(t *testing.T) {
	sub := decimal.RequireFromString("178.00")
	got := AddInvoiceSurcharge(sub, 6)
	want := decimal.RequireFromString("188.68")
	if !got.Equal(want) {
		t.Fatalf("got %s want %s", got, want)
	}
}
