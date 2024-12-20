package pkg

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Println("pgxpool.New ", err)
		return nil, err
	}
	err = db.Ping(ctx)
	if err != nil {
		log.Println("db.Ping ", err)
		return nil, err
	}
	return db, nil
}

func CloseDB(db *pgxpool.Pool) {
	db.Close()
}
