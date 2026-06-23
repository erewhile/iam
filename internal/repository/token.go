package repository

import (
	"context"

	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/ent/db/token"
	"github.com/erewhile/iam/pkg/utils"
	"github.com/google/uuid"
)

type TokenRepository interface {
	Create(ctx context.Context, params req.TokenCreate) error
	GetByID(ctx context.Context, id int) (*db.Token, error)
	RevokeByID(ctx context.Context, id int) error
	RevokeBySession(ctx context.Context, sessionID uuid.UUID) error
	RevokeByJTI(ctx context.Context, jti uuid.UUID) error
}

type tokenRepository struct {
	*baseRepository
}

var _ TokenRepository = (*tokenRepository)(nil)

func NewTokenRepository(client *db.Client) TokenRepository {
	return &tokenRepository{newBaseRepository(client)}
}

func (r *tokenRepository) Create(ctx context.Context, params req.TokenCreate) error {
	_, err := r.client.Token.Create().
		SetUserID(params.UserID).
		SetJti(params.JTI).
		SetSessionID(params.SessionID).
		SetType(params.Type).
		SetTokenHash(params.TokenHash).
		SetIP(params.IP).
		SetUserAgent(params.UserAgent).
		SetExpiresAt(params.ExpiresAt).
		Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *tokenRepository) GetByID(ctx context.Context, id int) (*db.Token, error) {
	t, err := r.client.Token.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *tokenRepository) RevokeByID(ctx context.Context, id int) error {
	_, err := r.client.Token.UpdateOneID(id).
		Where(token.RevokedAtIsNil(), token.ExpiresAtGT(utils.Now())).
		SetRevokedAt(utils.Now()).
		Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *tokenRepository) RevokeBySession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.client.Token.Update().
		Where(token.SessionIDEQ(sessionID), token.RevokedAtIsNil(), token.ExpiresAtGT(utils.Now())).
		SetRevokedAt(utils.Now()).
		Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *tokenRepository) RevokeByJTI(ctx context.Context, jti uuid.UUID) error {
	_, err := r.client.Token.Update().
		Where(token.JtiEQ(jti), token.RevokedAtIsNil(), token.ExpiresAtGT(utils.Now())).
		SetRevokedAt(utils.Now()).
		Save(ctx)
	if err != nil {
		return err
	}
	return nil
}
