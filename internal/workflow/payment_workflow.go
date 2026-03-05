package workflow

import (
	"errors"
	"time"

	"github.com/govalues/decimal"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const TaskQueue = "payment-task-queue"

type PaymentWorkflowInput struct {
	PaymentID  string
	ProviderID int
	AgentID    string
	RefID      string
	Amount     decimal.Decimal
	Gate       string
}

func PaymentWorkflow(ctx workflow.Context, input PaymentWorkflowInput) error {
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:        time.Second,
		BackoffCoefficient:     2.0,
		MaximumInterval:        30 * time.Second,
		MaximumAttempts:        10,
		NonRetryableErrorTypes: []string{"TerminalProviderError"},
	}

	providerActCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 45 * time.Second,
		RetryPolicy:         retryPolicy,
	})

	dbActCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 5,
		},
	})

	// Step 1: Call provider (Temporal retries on retryable errors)
	var providerResult ProviderResult
	providerErr := workflow.ExecuteActivity(providerActCtx, "CallProviderActivity", input).Get(providerActCtx, &providerResult)

	// Step 2: Update payment status regardless of outcome.
	// Extract error code from ApplicationError details when activity failed.
	errCode := providerResult.ErrorCode
	if providerErr != nil {
		var appErr *temporal.ApplicationError
		if errors.As(providerErr, &appErr) {
			_ = appErr.Details(&errCode)
		}
	}
	statusInput := UpdateStatusInput{
		PaymentID: input.PaymentID,
		Success:   providerErr == nil,
		ErrorCode: errCode,
	}
	_ = workflow.ExecuteActivity(dbActCtx, "UpdatePaymentStatusActivity", statusInput).Get(dbActCtx, nil)

	if providerErr != nil {
		return providerErr
	}

	// Steps 3 & 4: Only on success
	balanceInput := BalanceLogInput{
		PaymentID:  input.PaymentID,
		ProviderID: input.ProviderID,
		AgentID:    input.AgentID,
		RefID:      input.RefID,
		Amount:     input.Amount,
		Fee:        providerResult.Fee,
	}
	_ = workflow.ExecuteActivity(dbActCtx, "InsertBalanceLogActivity", balanceInput).Get(dbActCtx, nil)

	return nil
}
