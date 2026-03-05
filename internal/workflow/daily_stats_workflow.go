package workflow

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	DailyStatsTaskQueue = "daily-stats-task-queue"
	aggregateInterval   = 3 * time.Minute
)

func DailyStatsWorkflow(ctx workflow.Context) error {
	actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	})

	to := workflow.Now(ctx).UTC().Truncate(aggregateInterval)
	from := to.Add(-aggregateInterval)

	if err := workflow.ExecuteActivity(actCtx, "AggregateDailyStatsActivity", from, to).Get(actCtx, nil); err != nil {
		return err
	}
	return nil
}
