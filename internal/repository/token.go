package repository

import (
	"context"
	"time"

	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/model"
	"github.com/google/uuid"
)

type CreateTokenParams struct {
	UserID    int
	JTI       uuid.UUID
	SessionID uuid.UUID
	Type      model.TokenType
	TokenHash []byte
	IP        string
	UserAgent string
	ExpiresAt time.Time
}

type TokenRepository interface {
	Create(ctx context.Context, client *db.Client, params CreateTokenParams) error
	RevokeBySession(ctx context.Context, client *db.Client, sessionID uuid.UUID) error
	RevokeByJTI(ctx context.Context, client *db.Client, jti uuid.UUID) error
}

type tokenRepository struct {
	*baseRepository
}

var _ TokenRepository = (*tokenRepository)(nil)

func NewTokenRepository(client *db.Client) TokenRepository {
	return &tokenRepository{newBaseRepository(client)}
}

func (r *tokenRepository) Create(ctx context.Context, client *db.Client, p CreateTokenParams) error {
	_, err := client.Token.Create().
		SetUserID(p.UserID).
		SetJti(p.JTI).
		SetSessionID(p.SessionID).
		SetType(p.Type).
		SetTokenHash(p.TokenHash).
		SetIP(p.IP).
		SetUserAgent(p.UserAgent).
		SetExpiresAt(p.ExpiresAt).
		Save(ctx)

	return err
}

func (r *tokenRepository) RevokeBySession(ctx context.Context, client *db.Client, sessionID uuid.UUID) error {
	_, err := client.Token.Update().
		Where().
		Where().
		Save(ctx)

	return err
}

func (r *tokenRepository) RevokeByJTI(ctx context.Context, client *db.Client, jti uuid.UUID) error {
	_, err := client.Token.Update().
		Where().
		Save(ctx)

	return err
}
