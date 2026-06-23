package service

import (
	"context"
	"errors"

	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/cache/rds"
	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/logger"
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
		if db.IsNotFound(err) {
			return nil, errors.New("user not found")
		}
		logger.Error("login failed", err)
		return nil, errors.New("login failed")
	}

	if userInfo.Status != model.UserStatusActive {
		return nil, errors.New("account is disabled")
	}

	ok, err := password.Validate(param.Password, string(userInfo.PasswordHash))
	if err != nil {
		logger.Error("password check failed", err)
		return nil, errors.New("password check failed, please try again later")
	}

	if !ok {
		return nil, errors.New("wrong password")
	}

	sessionID := uuid.New()
	return s.issueTokenPair(ctx, userInfo.ID, userInfo.UUID, sessionID, param.RequestMeta)
}

func (s *UserService) LoginWithOAuthCode(ctx context.Context, payload *rds.OAuthCodePayload, meta req.RequestMeta) (*token.TokenPair, error) {
	return s.issueTokenPair(ctx, payload.UserID, payload.UserUUID, payload.SessionID, meta)
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

	tokenCache := rds.NewTokenCache()
	online, err := tokenCache.ExistsRefresh(ctx, claims.SessionID)
	if err != nil || !online {
		return nil, errors.New("refresh token expired")
	}

	_ = tokenCache.DelAccess(ctx, claims.SessionID)
	_ = tokenCache.DelRefresh(ctx, claims.SessionID)

	if err := s.token.RevokeBySession(ctx, claims.SessionID); err != nil {
		logger.Error("revoke failed", err)
		return nil, errors.New("revoke failed")
	}

	newSessionID := uuid.New()
	return s.issueTokenPair(ctx, payload.UserID, payload.UserUUID, newSessionID, param.RequestMeta)
}

func (s *UserService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	tokenCache := rds.NewTokenCache()
	_ = tokenCache.DelAccess(ctx, sessionID)
	_ = tokenCache.DelRefresh(ctx, sessionID)

	if err := s.token.RevokeBySession(ctx, sessionID); err != nil {
		logger.Error("logout failed", err)
		return errors.New("logout failed")
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
		logger.Error("generate token failed", err)
		return nil, errors.New("generate token failed")
	}

	accessJti := uuid.New()
	refreshJti := uuid.New()
	now := utils.Now()
	tokenCfg := config.Get().Token

	err = s.transactor.WithTx(ctx, func(ctx context.Context, txClient *db.Client) error {
		txTokenRepo := repository.NewTokenRepository(txClient)

		if err := txTokenRepo.Create(ctx, req.TokenCreate{
			UserID:    userID,
			Jti:       accessJti,
			SessionID: sessionID,
			Type:      model.TokenTypeAccess,
			TokenHash: hash.HashBlake2b256([]byte(tokenPair.AccessToken)),
			ExpiresAt: now.Add(tokenCfg.AccessTokenTTL),
			IP:        meta.IP,
			UserAgent: meta.UserAgent,
		}); err != nil {
			return err
		}

		return txTokenRepo.Create(ctx, req.TokenCreate{
			UserID:    userID,
			Jti:       refreshJti,
			SessionID: sessionID,
			Type:      model.TokenTypeRefresh,
			TokenHash: hash.HashBlake2b256([]byte(tokenPair.RefreshToken)),
			ExpiresAt: now.Add(tokenCfg.RefreshTokenTTL),
			IP:        meta.IP,
			UserAgent: meta.UserAgent,
		})
	})
	if err != nil {
		logger.Error("save token failed", err)
		return nil, errors.New("save token failed")
	}

	tokenCache := rds.NewTokenCache()
	_ = tokenCache.SetAccess(ctx, sessionID, config.Get().Token.AccessTokenTTL)
	_ = tokenCache.SetRefresh(ctx, sessionID, config.Get().Token.RefreshTokenTTL)

	return tokenPair, nil
}

func (s *UserService) List(ctx context.Context, params req.UserList) ([]resp.UserListItem, int, error) {
	content, count, err := s.repo.List(ctx, params)
	if err != nil {
		logger.Error("failed to retrieve the list", err)
		return nil, 0, errors.New("failed to retrieve the list")
	}

	return content, count, nil
}

func (s *UserService) Info(ctx context.Context, params req.InfoPathParams) (*resp.UserInfo, error) {
	userInfo, err := s.repo.GetByID(ctx, params.ID)

	if err != nil {
		if db.IsNotFound(err) {
			return nil, errors.New("user not found")
		}
		logger.Error("failed to get user info", err)
		return nil, errors.New("failed to get user info")
	}

	if userInfo.Status != model.UserStatusActive {
		return nil, errors.New("account is disabled")
	}

	return &resp.UserInfo{
		ID:           userInfo.ID,
		Username:     userInfo.Username,
		Email:        userInfo.Email,
		UUID:         userInfo.UUID,
		StatusDetail: userInfo.Status.String(),
	}, nil
}

func (s *UserService) Create(ctx context.Context, body req.UserCreate) error {
	if !body.Status.IsValid() {
		return errors.New("invalid user status")
	}

	exists, err := s.repo.Duplicate(ctx, body.Username, body.Email)
	if err != nil {
		logger.Error("failed to check if user exists", err)
		return errors.New("failed to check if user exists")
	}

	if exists {
		return errors.New("username or email already exists")
	}

	hashed, err := password.Hash(body.Password)
	if err != nil {
		logger.Error("failed to hash password", err)
		return errors.New("failed to hash password")
	}

	_, err = s.repo.Create(ctx, body, hashed)
	if err != nil {
		logger.Error("failed to create user", err)
		return errors.New("failed to create user")
	}
	return nil
}

func (s *UserService) Update(ctx context.Context, params req.UserUpdatePathParams, body req.UserUpdate) error {
	if !body.Status.IsValid() {
		return errors.New("invalid user status")
	}

	if body.Password != "" && len(body.Password) < 6 {
		// return errors.New("password must be greater than 6 characters")
		return errors.New("password must be at least 6 characters long")
	}

	_, err := s.repo.GetByID(ctx, params.UserID)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("user not found")
		}
		logger.Error("get user failed", err)
		return errors.New("failed to get user info")
	}

	exists, err := s.repo.Duplicate(ctx, body.Username, body.Email, params.UserID)
	if err != nil {
		logger.Error("failed to check if user exists", err)
		return errors.New("failed to check if user exists")
	}
	if exists {
		return errors.New("username or email already exists")
	}

	var hashed string
	if body.Password != "" {
		hashed, err = password.Hash(body.Password)
		if err != nil {
			logger.Error("failed to hash password", err)
			return errors.New("failed to hash password")
		}
	}

	_, err = s.repo.Update(ctx, params, body, hashed)
	if err != nil {
		logger.Error("failed to update user", err)
		return errors.New("failed to update user")
	}

	return nil
}

func (s *UserService) Delete(ctx context.Context, params req.DeletePathParams) error {
	_, err := s.repo.GetByID(ctx, params.ID)
	if err != nil {
		if db.IsNotFound(err) {
			return errors.New("user not found")

		}
		logger.Error("get user failed", err)
		return errors.New("failed to get user info")
	}

	if err := s.repo.Delete(ctx, params); err != nil {
		logger.Error("failed to delete user", err)
		return errors.New("failed to delete user")
	}
	return nil
}
