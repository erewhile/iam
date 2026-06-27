package repository

import (
	"context"

	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/ent/db/token"
	"github.com/erewhile/iam/internal/model"
	"github.com/erewhile/iam/pkg/utils"
	"github.com/google/uuid"
)

type TokenRepository interface {
	List(ctx context.Context, params req.TokenList) ([]resp.TokenListItem, int, error)
	Create(ctx context.Context, params req.TokenCreate) error
	GetByID(ctx context.Context, id int) (*db.Token, error)
	GetIfValid(ctx context.Context, hashed []byte, tokenType model.TokenType) (*db.Token, error)
	GetBySession(ctx context.Context, sessionID uuid.UUID) (*db.Token, error)
	RevokeByID(ctx context.Context, id int) error
	RevokeBySession(ctx context.Context, sessionID uuid.UUID) error
	RevokeByJTI(ctx context.Context, jti uuid.UUID) error
	RevokeAllByUser(ctx context.Context, userID int) error
	ListActiveSessionsByUser(ctx context.Context, userID int) ([]uuid.UUID, error)
	ClearExpiredOrRevoked(ctx context.Context) (int, error)
}

type tokenRepository struct {
	*baseRepository
}

var _ TokenRepository = (*tokenRepository)(nil)

func NewTokenRepository(client *db.Client) TokenRepository {
	return &tokenRepository{newBaseRepository(client)}
}

func (r *tokenRepository) List(ctx context.Context, params req.TokenList) ([]resp.TokenListItem, int, error) {
	q := r.client.Token.Query().
		Where(
			token.ExpiresAtGT(utils.Now()),
			token.RevokedAtIsNil(),
		)

	if params.UserID > 0 {
		q = q.Where(token.UserIDEQ(params.UserID))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []resp.TokenListItem{}, 0, nil
	}

	totalPages := (total + params.PerPage - 1) / params.PerPage
	if params.Page > totalPages {
		return []resp.TokenListItem{}, 0, nil
	}

	offset := (params.Page - 1) * params.PerPage

	tokens, err := q.
		Order(db.Desc(token.FieldID)).
		Offset(offset).
		Limit(params.PerPage).
		All(ctx)
	if err != nil {
		return []resp.TokenListItem{}, 0, err
	}

	result := make([]resp.TokenListItem, 0, len(tokens))
	for _, item := range tokens {
		result = append(result, resp.TokenListItem{
			ID:         item.ID,
			UserID:     item.UserID,
			Jti:        item.Jti,
			SessionID:  item.SessionID,
			TypeDetail: item.Type.String(),
			IP:         item.IP,
			UserAgent:  item.UserAgent,
			ExpiresAt:  item.ExpiresAt,
		})
	}
	return result, total, nil
}

func (r *tokenRepository) Create(ctx context.Context, params req.TokenCreate) error {
	_, err := r.client.Token.Create().
		SetUserID(params.UserID).
		SetJti(params.Jti).
		SetCookieID(params.CookieID).
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
	t, err := r.client.Token.Query().
		Where(
			token.IDEQ(id),
			token.ExpiresAtGT(utils.Now()),
			token.RevokedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *tokenRepository) GetIfValid(ctx context.Context, hashed []byte, tokenType model.TokenType) (*db.Token, error) {
	t, err := r.client.Token.Query().
		Where(
			token.TokenHashEQ(hashed),
			token.TypeEQ(tokenType),
			token.ExpiresAtGT(utils.Now()),
			token.RevokedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *tokenRepository) GetBySession(ctx context.Context, sessionID uuid.UUID) (*db.Token, error) {
	t, err := r.client.Token.Query().
		Where(
			token.SessionIDEQ(sessionID),
			token.ExpiresAtGT(utils.Now()),
			token.RevokedAtIsNil(),
		).
		First(ctx)
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

func (r *tokenRepository) RevokeAllByUser(ctx context.Context, userID int) error {
	_, err := r.client.Token.Update().
		Where(
			token.UserIDEQ(userID),
			token.RevokedAtIsNil(),
			token.ExpiresAtGT(utils.Now()),
		).
		SetRevokedAt(utils.Now()).
		Save(ctx)
	return err
}

func (r *tokenRepository) ListActiveSessionsByUser(ctx context.Context, userID int) ([]uuid.UUID, error) {
	var sessionIDs []uuid.UUID
	err := r.client.Token.Query().
		Where(
			token.UserIDEQ(userID),
			token.RevokedAtIsNil(),
			token.ExpiresAtGT(utils.Now()),
		).
		Select(token.FieldSessionID).
		Scan(ctx, &sessionIDs)
	if err != nil {
		return nil, err
	}
	return sessionIDs, nil
}

func (r *tokenRepository) ClearExpiredOrRevoked(ctx context.Context) (int, error) {
	now := utils.Now()
	batchSize := 1000

	ids, err := r.client.Token.Query().
		Where(
			token.Or(
				token.ExpiresAtLTE(now),
				token.RevokedAtNotNil(),
			),
		).
		Limit(batchSize).
		Select(token.FieldID).
		Ints(ctx)

	if err != nil {
		return 0, err
	}

	if len(ids) == 0 {
		return 0, nil
	}

	affected, err := r.client.Token.Delete().
		Where(token.IDIn(ids...)).
		Exec(ctx)

	if err != nil {
		return 0, err
	}

	return affected, nil
}
