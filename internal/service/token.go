package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/erewhile/iam/internal/cache/rds"
	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/logger"
	"github.com/erewhile/iam/internal/repository"
)

type TokenService struct {
	repo         repository.TokenRepository
	userRepo     repository.UserRepository
	tokenCache   rds.TokenCache
	sessionCache rds.IAMSessionCache
}

func NewTokenService(
	repo repository.TokenRepository,
	tokenCache rds.TokenCache,
	userRepo repository.UserRepository,
	sessionCache rds.IAMSessionCache,
) *TokenService {
	return &TokenService{
		repo:         repo,
		tokenCache:   tokenCache,
		userRepo:     userRepo,
		sessionCache: sessionCache,
	}
}

func (s *TokenService) List(ctx context.Context, params req.TokenList) ([]resp.TokenListItem, int, error) {
	content, count, err := s.repo.List(ctx, params)
	if err != nil {
		logger.Error("failed to retrieve the list", err)
		return nil, 0, errors.New("failed to retrieve the list")
	}

	users, err := s.userRepo.GetAll(ctx)
	if err != nil {
		logger.Error("failed to get all users for token mapping", err)
	}

	userMap := make(map[int]string)
	for _, u := range users {
		userMap[u.ID] = u.Username
	}

	listItems := make([]resp.TokenListItem, len(content))
	for i, t := range content {
		username := userMap[t.UserID]
		if username == "" {
			username = "Unknown"
		}

		listItems[i] = resp.TokenListItem{
			ID:         t.ID,
			UserID:     t.UserID,
			Username:   username,
			TypeDetail: t.TypeDetail,
			IP:         t.IP,
			Jti:        t.Jti,
			UserAgent:  t.UserAgent,
			ExpiresAt:  t.ExpiresAt,
		}
	}

	return listItems, count, nil
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

func (s *TokenService) Revoke(ctx context.Context, params req.TokenRevokePathParams) error {
	tokenInfo, err := s.repo.GetByID(ctx, params.TokenID)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("token not found")
		}
		logger.Error("failed to get token info", err.Error())
		return errors.New("failed to get token info")
	}

	var errs []error

	if err := s.tokenCache.DelAccess(ctx, tokenInfo.SessionID); err != nil {
		errs = append(errs, fmt.Errorf("del access token failed (session=%s): %w", tokenInfo.SessionID, err))
	}
	if err := s.tokenCache.DelRefresh(ctx, tokenInfo.SessionID); err != nil {
		errs = append(errs, fmt.Errorf("del refresh token failed (session=%s): %w", tokenInfo.SessionID, err))
	}
	if tokenInfo.CookieID != "" {
		if err := s.sessionCache.Del(ctx, tokenInfo.CookieID); err != nil {
			errs = append(errs, fmt.Errorf("clear iam session failed (session=%s): %w", tokenInfo.SessionID, err))
		}
	}

	if err := s.repo.RevokeBySession(ctx, tokenInfo.SessionID); err != nil {
		errs = append(errs, fmt.Errorf("revoke by session failed (session=%s): %w", tokenInfo.SessionID, err))
	}

	if len(errs) > 0 {
		err := errors.Join(errs...)
		logger.Error("failed to revoke token", err.Error())
		return errors.New("failed to revoke token")
	}

	logger.Info("token revoked", "sessionID", tokenInfo.SessionID)
	return nil
}
