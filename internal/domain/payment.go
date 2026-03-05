package domain

import (
	"time"

	"github.com/govalues/decimal"
)

type PaymentStatus int16

const (
	StatusPending  PaymentStatus = 1
	StatusSuccess  PaymentStatus = 2
	StatusFailed   PaymentStatus = 3
	StatusCanceled PaymentStatus = 4
)

type ErrorCode int16

const (
	ErrCodeNone             ErrorCode = 0
	ErrCodeInsufficientFunds ErrorCode = 101
	ErrCodeInvalidAccount   ErrorCode = 102
	ErrCodeTimeout          ErrorCode = 201
	ErrCodeProviderDown     ErrorCode = 202
)

// TerminalErrorCodes are error codes that must NOT be retried.
var TerminalErrorCodes = map[ErrorCode]bool{
	ErrCodeInsufficientFunds: true,
	ErrCodeInvalidAccount:   true,
}

type Payment struct {
	ID         string
	ProviderID int
	AgentID    string
	RefID      string
	Amount     decimal.Decimal
	Fee        decimal.Decimal
	Status     PaymentStatus
	ErrorCode  ErrorCode
	PayDate    time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	CanceledAt *time.Time
}

type BalanceLog struct {
	LogDate      time.Time
	PartnerID    int
	ProviderID   int
	Amount       decimal.Decimal
	RefID        string
	Type         string // payment, topup, cancel, fix
	ProviderName string
	PartnerName  string
}
