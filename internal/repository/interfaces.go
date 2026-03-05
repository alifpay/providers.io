package repository

import (
	"context"
	"time"

	"github.com/alifpay/providers.io/internal/domain"
)

type PaymentRepository interface {
	Insert(ctx context.Context, p *domain.Payment) (string, error)
	UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus, errCode domain.ErrorCode) error
	GetByID(ctx context.Context, id string) (*domain.Payment, error)
}

type BalanceLogRepository interface {
	Insert(ctx context.Context, log *domain.BalanceLog) error
}

type ProviderRepository interface {
	GetByID(ctx context.Context, id int) (*domain.Provider, *domain.Partner, error)
}

type DailyStatsRepository interface {
	Aggregate(ctx context.Context, from, to time.Time) error
}
