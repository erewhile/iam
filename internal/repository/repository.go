package repository

import (
	"context"
	"fmt"

	"github.com/erewhile/iam/internal/ent/db"
)

type baseRepository struct {
	client *db.Client
}

func newBaseRepository(client *db.Client) *baseRepository {
	return &baseRepository{client: client}
}

func (r *baseRepository) Client() *db.Client {
	return r.client
}

type txKey struct{}

func txFromContext(ctx context.Context) (*db.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*db.Tx)
	return tx, ok
}

type Transactor struct {
	client *db.Client
}

func NewTransactor(client *db.Client) *Transactor {
	return &Transactor{client: client}
}

func (t *Transactor) WithTx(ctx context.Context, fn func(ctx context.Context, txClient *db.Client) error) error {
	if tx, ok := txFromContext(ctx); ok {
		return fn(ctx, tx.Client())
	}

	tx, err := t.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	defer func() {
		if p := recover(); p != nil {
			if rerr := tx.Rollback(); rerr != nil {
				fmt.Printf("rollback after panic failed: %v (panic: %v)\n", rerr, p)
			}
			panic(p)
		}
	}()

	if err := fn(txCtx, tx.Client()); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("rolling back transaction: %v (original error: %w)", rerr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
