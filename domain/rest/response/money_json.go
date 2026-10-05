package response

import (
	"github.com/shopspring/decimal"
)

// Money 对外 JSON 金额（元，保留两位小数）。
type Money decimal.Decimal

func MoneyFrom(d decimal.Decimal) Money {
	return Money(d.Round(2))
}

func (m Money) MarshalJSON() ([]byte, error) {
	return decimal.Decimal(m).MarshalJSON()
}
