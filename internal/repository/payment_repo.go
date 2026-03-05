package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alifpay/providers.io/internal/domain"
)

type paymentRepo struct {
	db *pgxpool.Pool
}

func NewPaymentRepo(db *pgxpool.Pool) PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Insert(ctx context.Context, p *domain.Payment) (string, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO payments (provider_id, agent_id, ref_id, amount, fee, status, pay_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (agent_id, ref_id, pay_date) DO NOTHING
		RETURNING id`,
		p.ProviderID, p.AgentID, p.RefID,
		p.Amount, p.Fee, p.Status, p.PayDate,
	).Scan(&id)
	return id, err
}

func (r *paymentRepo) UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus, errCode domain.ErrorCode) error {
	_, err := r.db.Exec(ctx, `
		UPDATE payments SET status = $2, error_code = $3, updated_at = NOW()
		WHERE id = $1`,
		id, status, errCode,
	)
	return err
}

func (r *paymentRepo) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, provider_id, agent_id, ref_id, amount, fee, status, error_code,
		       pay_date, created_at, updated_at, canceled_at
		FROM payments WHERE id = $1`, id)

	p := &domain.Payment{}

	err := row.Scan(
		&p.ID, &p.ProviderID, &p.AgentID, &p.RefID,
		&p.Amount, &p.Fee, &p.Status, &p.ErrorCode,
		&p.PayDate, &p.CreatedAt, &p.UpdatedAt, &p.CanceledAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("payment not found: %s", id)
		}
		return nil, err
	}

	return p, nil
}
