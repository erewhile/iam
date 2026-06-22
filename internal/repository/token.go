package repository

import (
	"context"

	"github.com/erewhile/iam/internal/ent/db"
)

type TokenRepository interface {
	Logout(ctx context.Context)
}

type tokenRepository struct {
	*baseRepository
}

var _ TokenRepository = (*tokenRepository)(nil)

func NewTokenRepository(client *db.Client) TokenRepository {
	return &tokenRepository{newBaseRepository(client)}
}

func (r *tokenRepository) Logout(ctx context.Context) {}
