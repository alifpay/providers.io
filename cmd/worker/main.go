package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/alifpay/providers.io/internal/config"
	"github.com/alifpay/providers.io/internal/provider"
	"github.com/alifpay/providers.io/internal/repository"
	wf "github.com/alifpay/providers.io/internal/workflow"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	tc, err := client.Dial(client.Options{
		HostPort:  cfg.TemporalHost,
		Namespace: cfg.TemporalNamespace,
	})
	if err != nil {
		slog.Error("connect to temporal", "err", err)
		os.Exit(1)
	}
	defer tc.Close()

	acts := &wf.Activities{
		PaymentRepo:    repository.NewPaymentRepo(pool),
		BalanceRepo:    repository.NewBalanceLogRepo(pool),
		ProviderRepo:   repository.NewProviderRepo(pool),
		ProviderClient: &provider.MockClient{},
	}

	statsActs := &wf.DailyStatsActivities{
		DailyStatsRepo: repository.NewDailyStatsRepo(pool),
	}

	w := worker.New(tc, wf.TaskQueue, worker.Options{})
	w.RegisterWorkflow(wf.PaymentWorkflow)
	w.RegisterActivity(acts)

	statsWorker := worker.New(tc, wf.DailyStatsTaskQueue, worker.Options{})
	statsWorker.RegisterWorkflow(wf.DailyStatsWorkflow)
	statsWorker.RegisterActivity(statsActs)

	if err := w.Start(); err != nil {
		slog.Error("payment worker start", "err", err)
		os.Exit(1)
	}
	defer w.Stop()

	slog.Info("workers starting", "payment_queue", wf.TaskQueue, "stats_queue", wf.DailyStatsTaskQueue)
	if err := statsWorker.Run(worker.InterruptCh()); err != nil {
		slog.Error("stats worker error", "err", err)
		os.Exit(1)
	}
}
