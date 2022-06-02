package store

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TxQuerier interface {
	Querier
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, Querier, error)
	ExecWithTx(context.Context, pgx.TxOptions, func(Querier) error) error
}

type TxQueries struct {
	*Queries
	pool *pgxpool.Pool
}

func (txq *TxQueries) BeginTx(ctx context.Context, txOpts pgx.TxOptions) (pgx.Tx, Querier, error) {
	tx, err := txq.pool.BeginTx(ctx, txOpts)
	queries := txq.Queries.WithTx(tx)
	return tx, queries, err
}

func (txq *TxQueries) ExecWithTx(ctx context.Context, txOpts pgx.TxOptions, task func(Querier) error) error {
	return txq.pool.BeginTxFunc(ctx, txOpts, func(t pgx.Tx) error {
		queries := txq.Queries.WithTx(t)
		return task(queries)
	})
}

func NewTxQueries(pool *pgxpool.Pool) *TxQueries {
	return &TxQueries{
		Queries: New(pool),
		pool:    pool,
	}
}
