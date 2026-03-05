package repository

import "github.com/govalues/decimal"

func ParseDecimal(s string) (decimal.Decimal, error) {
	return decimal.Parse(s)
}

func parseDecimal(s string) (decimal.Decimal, error) {
	return ParseDecimal(s)
}
