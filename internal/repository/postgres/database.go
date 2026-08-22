package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QuerySession interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Database struct {
	Pool               *pgxpool.Pool
	transactionTimeout time.Duration
}

func Open(ctx context.Context, url string) (*Database, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	config.MinConns = 1
	config.MaxConns = 12
	config.MaxConnIdleTime = 5 * time.Minute
	config.MaxConnLifetime = 30 * time.Minute
	config.HealthCheckPeriod = 20 * time.Second
	config.ConnConfig.RuntimeParams["application_name"] = "cry076_resume_consistency"
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Database{Pool: pool, transactionTimeout: 8 * time.Second}, nil
}
func (d *Database) Close()                          { d.Pool.Close() }
func (d *Database) Ready(ctx context.Context) error { return d.Pool.Ping(ctx) }

func (d *Database) WithinTransaction(ctx context.Context, operation func(context.Context) error) error {
	if _, alreadyAtomic := ctx.Value(resumeTransactionKey{}).(pgx.Tx); alreadyAtomic {
		return operation(ctx)
	}
	txCtx, cancel := context.WithTimeout(ctx, d.transactionTimeout)
	defer cancel()
	tx, err := d.Pool.BeginTx(txCtx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return fmt.Errorf("begin material consistency transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(txCtx)) }()

	workCtx := context.WithValue(txCtx, resumeTransactionKey{}, tx)
	if err := operation(workCtx); err != nil {
		return fmt.Errorf("apply material consistency operation: %w", err)
	}
	if err := txCtx.Err(); err != nil {
		return fmt.Errorf("material consistency transaction deadline: %w", err)
	}
	if err := tx.Commit(txCtx); err != nil {
		if errors.Is(err, pgx.ErrTxCommitRollback) {
			return fmt.Errorf("material consistency serialization conflict: %w", err)
		}
		return fmt.Errorf("commit material consistency transaction: %w", err)
	}
	return nil
}

type resumeTransactionKey struct{}

func (d *Database) queryer(ctx context.Context) QuerySession {
	if tx, ok := ctx.Value(resumeTransactionKey{}).(pgx.Tx); ok {
		return tx
	}
	return d.Pool
}
