package workflow

import (
	"context"
	"time"

	"go.temporal.io/sdk/temporal"

	"github.com/alifpay/providers.io/internal/domain"
	"github.com/alifpay/providers.io/internal/provider"
	"github.com/alifpay/providers.io/internal/repository"
	"github.com/govalues/decimal"
)

// Activities holds dependencies for payment workflow activities.
type Activities struct {
	PaymentRepo    repository.PaymentRepository
	BalanceRepo    repository.BalanceLogRepository
	ProviderRepo   repository.ProviderRepository
	ProviderClient provider.Client
}

// DailyStatsActivities holds dependencies for daily stats workflow activities.
type DailyStatsActivities struct {
	DailyStatsRepo repository.DailyStatsRepository
}

type ProviderResult struct {
	Success   bool
	ErrorCode domain.ErrorCode
	Fee       decimal.Decimal
}

type UpdateStatusInput struct {
	PaymentID string
	Success   bool
	ErrorCode domain.ErrorCode
}

type BalanceLogInput struct {
	PaymentID  string
	ProviderID int
	AgentID    string
	RefID      string
	Amount     decimal.Decimal
	Fee        decimal.Decimal
}

func (a *Activities) CallProviderActivity(ctx context.Context, input PaymentWorkflowInput) (ProviderResult, error) {
	resp, err := a.ProviderClient.Send(ctx, input.PaymentID, input.Amount, input.Gate)
	if err != nil {
		return ProviderResult{}, err
	}

	if !resp.Success {
		if domain.TerminalErrorCodes[resp.ErrorCode] {
			return ProviderResult{},
				temporal.NewNonRetryableApplicationError(
					"terminal provider error",
					"TerminalProviderError",
					nil,
					resp.ErrorCode,
				)
		}
		return ProviderResult{},
			temporal.NewApplicationError(
				"provider temporary failure",
				"RetryableProviderError",
				nil,
				resp.ErrorCode,
			)
	}

	return ProviderResult{Success: true, Fee: resp.Fee}, nil
}

func (a *Activities) UpdatePaymentStatusActivity(ctx context.Context, input UpdateStatusInput) error {
	status := domain.StatusFailed
	if input.Success {
		status = domain.StatusSuccess
	}
	return a.PaymentRepo.UpdateStatus(ctx, input.PaymentID, status, input.ErrorCode)
}

func (a *Activities) InsertBalanceLogActivity(ctx context.Context, input BalanceLogInput) error {
	prov, part, err := a.ProviderRepo.GetByID(ctx, input.ProviderID)
	if err != nil {
		return err
	}

	log := &domain.BalanceLog{
		PartnerID:    part.ID,
		ProviderID:   prov.ID,
		Amount:       input.Amount,
		RefID:        input.RefID,
		Type:         "payment",
		ProviderName: prov.Name,
		PartnerName:  part.Name,
	}
	return a.BalanceRepo.Insert(ctx, log)
}

func (a *DailyStatsActivities) AggregateDailyStatsActivity(ctx context.Context, from, to time.Time) error {
	return a.DailyStatsRepo.Aggregate(ctx, from, to)
}
