package money

import (
	"github.com/shopspring/decimal"
)

const Scale = 2

// Decimal 金额（元），与库 DECIMAL(16,2) 对应。
type Decimal = decimal.Decimal

var (
	Zero = decimal.Zero
)

func FromInt(yuan int64) Decimal {
	return decimal.NewFromInt(yuan)
}

func MulQty(unit Decimal, qty int) Decimal {
	if qty <= 0 {
		return Zero
	}
	return unit.Mul(decimal.NewFromInt(int64(qty))).Round(Scale)
}

// AddInvoiceSurcharge 企业开票加价（如 6%）。
func AddInvoiceSurcharge(subtotal Decimal, percent int) Decimal {
	if percent <= 0 {
		return subtotal.Round(Scale)
	}
	pct := decimal.NewFromInt(int64(percent)).Div(decimal.NewFromInt(100))
	return subtotal.Add(subtotal.Mul(pct)).Round(Scale)
}

func Neg(v Decimal) Decimal {
	return v.Neg()
}
