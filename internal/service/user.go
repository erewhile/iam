package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/model"
	"github.com/erewhile/iam/internal/repository"
	"github.com/erewhile/iam/internal/token"
	"github.com/erewhile/iam/pkg/hash"
	"github.com/erewhile/iam/pkg/password"
	"github.com/erewhile/iam/pkg/utils"
	"github.com/google/uuid"
)

type UserService struct {
	repo       repository.UserRepository
	token      repository.TokenRepository
	transactor *repository.Transactor
}

func NewUserService(
	repo repository.UserRepository,
	token repository.TokenRepository,
	transactor *repository.Transactor,
) *UserService {
	return &UserService{
		repo:       repo,
		token:      token,
		transactor: transactor,
	}
}

func (s *UserService) Login(ctx context.Context, param req.UserLogin) (*token.TokenPair, error) {
	userInfo, err := s.repo.GetByUsername(ctx, param.Username)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	if userInfo == nil {
		return nil, errors.New("user not found")
	}

	if userInfo.Status != model.UserStatusActive {
		return nil, errors.New("account is disabled")
	}

	ok, err := password.Validate(param.Password, string(userInfo.PasswordHash))
	if err != nil {
		return nil, errors.New("password check failed, please try again later")
	}

	if !ok {
		return nil, errors.New("wrong password")
	}

	sessionID := uuid.New()
	return s.issueTokenPair(ctx, userInfo.ID, userInfo.UUID, sessionID, param.RequestMeta)
}

func (s *UserService) Profile(ctx context.Context, userID int) {}

func (s *UserService) Refresh(ctx context.Context, param req.UserRefresh) (*token.TokenPair, error) {
	claims, payload, err := token.Validate(
		param.Token,
		[]byte(config.Get().Token.Aad),
		token.TokenTypeRefresh,
	)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	if err := s.token.RevokeBySession(ctx, claims.SessionID); err != nil {
		return nil, errors.New("revoke failed")
	}

	newSessionID := uuid.New()
	return s.issueTokenPair(ctx, payload.UserID, payload.UserUUID, newSessionID, param.RequestMeta)
}

func (s *UserService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.token.RevokeBySession(ctx, sessionID); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}
	return nil
}

func (s *UserService) issueTokenPair(
	ctx context.Context,
	userID int,
	userUUID uuid.UUID,
	sessionID uuid.UUID,
	meta req.RequestMeta,
) (*token.TokenPair, error) {
	tokenPair, err := token.Generate(
		userID,
		userUUID,
		sessionID,
		[]byte(config.Get().Token.Aad),
	)
	if err != nil {
		return nil, fmt.Errorf("generate token failed: %w", err)
	}

	accessJti := uuid.New()
	refreshJti := uuid.New()
	now := utils.Now()
	tokenCfg := config.Get().Token

	err = s.transactor.WithTx(ctx, func(ctx context.Context, txClient *db.Client) error {
		txTokenRepo := repository.NewTokenRepository(txClient)

		if err := txTokenRepo.Create(ctx, repository.CreateTokenParams{
			UserID:    userID,
			JTI:       accessJti,
			SessionID: sessionID,
			Type:      model.TokenTypeAccess,
			TokenHash: hash.HashBlake2b256([]byte(tokenPair.AccessToken)),
			ExpiresAt: now.Add(tokenCfg.AccessTokenTTL),
			IP:        meta.IP,
			UserAgent: meta.UserAgent,
		}); err != nil {
			return err
		}

		return txTokenRepo.Create(ctx, repository.CreateTokenParams{
			UserID:    userID,
			JTI:       refreshJti,
			SessionID: sessionID,
			Type:      model.TokenTypeRefresh,
			TokenHash: hash.HashBlake2b256([]byte(tokenPair.RefreshToken)),
			ExpiresAt: now.Add(tokenCfg.RefreshTokenTTL),
			IP:        meta.IP,
			UserAgent: meta.UserAgent,
		})
	})
	if err != nil {
		return nil, fmt.Errorf("save token failed: %w", err)
	}

	return tokenPair, nil
}
