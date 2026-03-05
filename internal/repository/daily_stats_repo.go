package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type dailyStatsRepo struct {
	db *pgxpool.Pool
}

func NewDailyStatsRepo(db *pgxpool.Pool) DailyStatsRepository {
	return &dailyStatsRepo{db: db}
}

// Aggregate upserts daily_stats for all providers that had successful payments
// with updated_at in [from, to).
func (r *dailyStatsRepo) Aggregate(ctx context.Context, from, to time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO daily_stats (stat_date, provider_id, total_pay, count_pay)
		SELECT
		    pay_date,
		    provider_id,
		    SUM(amount),
		    COUNT(*)
		FROM payments
		WHERE status = 2
		  AND updated_at >= $1
		  AND updated_at < $2
		GROUP BY pay_date, provider_id
		ON CONFLICT (stat_date, provider_id) DO UPDATE SET
		    total_pay = daily_stats.total_pay + EXCLUDED.total_pay,
		    count_pay = daily_stats.count_pay + EXCLUDED.count_pay`,
		from, to,
	)
	return err
}
