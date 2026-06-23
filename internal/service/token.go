package service

import (
	"context"
	"errors"

	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/logger"
	"github.com/erewhile/iam/internal/repository"
)

type TokenService struct {
	repo repository.TokenRepository
}

func NewTokenService(repo repository.TokenRepository) *TokenService {
	return &TokenService{repo}
}

func (s *TokenService) List(ctx context.Context, params req.TokenList) ([]resp.TokenListItem, int, error) {
	content, count, err := s.repo.List(ctx, params)
	if err != nil {
		logger.Error("failed to retrieve the list", err)
		return nil, 0, errors.New("failed to retrieve the list")
	}

	return content, count, nil
}

func (s *TokenService) Info(ctx context.Context, params req.InfoPathParams) (*resp.TokenInfo, error) {
	tokenInfo, err := s.repo.GetByID(ctx, params.ID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errors.New("token not found")
		}
		logger.Error("failed to get token info", err)
		return nil, errors.New("failed to get token info")
	}

	return &resp.TokenInfo{
		ID:         tokenInfo.ID,
		UserID:     tokenInfo.UserID,
		Jti:        tokenInfo.Jti,
		SessionID:  tokenInfo.SessionID,
		TypeDetail: tokenInfo.Type.String(),
		IP:         tokenInfo.IP,
		UserAgent:  tokenInfo.UserAgent,
		ExpiresAt:  tokenInfo.ExpiresAt,
	}, nil
}

func (s *TokenService) Revoke(ctx context.Context, params req.TokenRevoke) error {
	tokenInfo, err := s.repo.GetByID(ctx, params.ID)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("token not found")
		}
		logger.Error("failed to get token info", err.Error())
		return errors.New("failed to get token info")
	}

	if err := s.repo.RevokeBySession(ctx, tokenInfo.SessionID); err != nil {
		logger.Error("failed to revoke token", err.Error())
		return errors.New("failed to revoke token")
	}
	return nil
}
