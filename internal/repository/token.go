package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/ent/db/token"
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
	Create(ctx context.Context, params CreateTokenParams) error
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

func (r *tokenRepository) Create(ctx context.Context, p CreateTokenParams) error {
	_, err := r.client.Token.Create().
		SetUserID(p.UserID).
		SetJti(p.JTI).
		SetSessionID(p.SessionID).
		SetType(p.Type).
		SetTokenHash(p.TokenHash).
		SetIP(p.IP).
		SetUserAgent(p.UserAgent).
		SetExpiresAt(p.ExpiresAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create token (jti=%s, session=%s): %w", p.JTI, p.SessionID, err)
	}
	return nil
}

func (r *tokenRepository) RevokeBySession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.client.Token.Update().
		Where(token.SessionIDEQ(sessionID), token.RevokedAtIsNil(), token.ExpiresAtGT(time.Now())).
		SetRevokedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("revoke session %s: %w", sessionID, err)
	}
	return nil
}

func (r *tokenRepository) RevokeByJTI(ctx context.Context, jti uuid.UUID) error {
	_, err := r.client.Token.Update().
		Where(token.JtiEQ(jti), token.RevokedAtIsNil(), token.ExpiresAtGT(time.Now())).
		SetRevokedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("revoke jti %s: %w", jti, err)
	}
	return nil
}
