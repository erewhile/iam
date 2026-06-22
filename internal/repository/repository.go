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

type Transactor struct {
	client *db.Client
}

func NewTransactor(client *db.Client) *Transactor {
	return &Transactor{client: client}
}

func (t *Transactor) WithTx(ctx context.Context, fn func(txClient *db.Client) error) error {
	tx, err := t.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx.Client()); err != nil {
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
