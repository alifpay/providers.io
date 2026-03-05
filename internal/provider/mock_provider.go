package provider

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/alifpay/providers.io/internal/domain"
	"github.com/govalues/decimal"
)

type ProviderResponse struct {
	Success   bool
	ErrorCode domain.ErrorCode
	Fee       decimal.Decimal
}

type Client interface {
	Send(ctx context.Context, paymentID string, amount decimal.Decimal, gate string) (*ProviderResponse, error)
}

// MockClient simulates an external payment provider.
// Outcomes: 70% success, 20% retryable error, 10% terminal error.
type MockClient struct{}

func (m *MockClient) Send(_ context.Context, _ string, _ decimal.Decimal, _ string) (*ProviderResponse, error) {
	time.Sleep(time.Duration(100+rand.IntN(400)) * time.Millisecond)

	fee, _ := decimal.Parse("0.50")
	n := rand.IntN(10)
	switch {
	case n < 7:
		return &ProviderResponse{Success: true, ErrorCode: domain.ErrCodeNone, Fee: fee}, nil
	case n < 9:
		return &ProviderResponse{Success: false, ErrorCode: domain.ErrCodeProviderDown}, nil
	default:
		return &ProviderResponse{Success: false, ErrorCode: domain.ErrCodeInsufficientFunds}, nil
	}
}
