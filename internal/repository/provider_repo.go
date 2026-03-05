package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alifpay/providers.io/internal/domain"
)

type providerRepo struct {
	db *pgxpool.Pool
}

func NewProviderRepo(db *pgxpool.Pool) ProviderRepository {
	return &providerRepo{db: db}
}

func (r *providerRepo) GetByID(ctx context.Context, id int) (*domain.Provider, *domain.Partner, error) {
	row := r.db.QueryRow(ctx, `
		SELECT p.id, p.partner_id, p.name, p.gate, p.currency, p.active,
		       p.min_amount, p.max_amount,
		       pa.id, pa.country, pa.name, pa.ref_id
		FROM providers p
		JOIN partners pa ON pa.id = p.partner_id
		WHERE p.id = $1`, id)

	prov := &domain.Provider{}
	part := &domain.Partner{}

	err := row.Scan(
		&prov.ID, &prov.PartnerID, &prov.Name, &prov.Gate,
		&prov.Currency, &prov.Active, &prov.MinAmount, &prov.MaxAmount,
		&part.ID, &part.Country, &part.Name, &part.RefID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, fmt.Errorf("provider not found: %d", id)
		}
		return nil, nil, err
	}

	return prov, part, nil
}
