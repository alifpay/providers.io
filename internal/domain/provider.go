package domain

import "github.com/govalues/decimal"

type Provider struct {
	ID        int
	PartnerID int
	Name      string
	Gate      string
	Currency  string
	Active    bool
	MinAmount decimal.Decimal
	MaxAmount decimal.Decimal
}

type Partner struct {
	ID      int
	Country string
	Name    string
	RefID   string
}
