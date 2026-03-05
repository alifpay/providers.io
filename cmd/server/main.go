package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.temporal.io/sdk/client"

	"github.com/alifpay/providers.io/internal/api"
	"github.com/alifpay/providers.io/internal/config"
	"github.com/alifpay/providers.io/internal/repository"
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

	h := &api.Handler{
		PaymentRepo:    repository.NewPaymentRepo(pool),
		ProviderRepo:   repository.NewProviderRepo(pool),
		TemporalClient: tc,
	}

	router := api.NewRouter(h)
	slog.Info("server starting", "addr", cfg.HTTPAddr)

	if err := http.ListenAndServe(cfg.HTTPAddr, router); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
