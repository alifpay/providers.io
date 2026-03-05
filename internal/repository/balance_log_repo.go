package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alifpay/providers.io/internal/domain"
)

type balanceLogRepo struct {
	db *pgxpool.Pool
}

func NewBalanceLogRepo(db *pgxpool.Pool) BalanceLogRepository {
	return &balanceLogRepo{db: db}
}

func (r *balanceLogRepo) Insert(ctx context.Context, log *domain.BalanceLog) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Lock by partner_id to serialize balance updates across all providers for this partner.
	// pg_advisory_xact_lock is released automatically when the transaction ends.
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, int64(log.PartnerID)); err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO balance_log
		    (partner_id, provider_id, amount, balance, ref_id, type, provider_name, partner_name)
		SELECT
		    $1, $2, $3,
		    COALESCE((SELECT balance FROM balance_log WHERE partner_id = $1 ORDER BY id DESC LIMIT 1), 0)
		        + CASE WHEN $5 = 'payment' THEN -$3::DECIMAL(18,2) ELSE $3::DECIMAL(18,2) END,
		    $4, $5, $6, $7`,
		log.PartnerID, log.ProviderID,
		log.Amount,
		log.RefID, log.Type, log.ProviderName, log.PartnerName,
	)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
