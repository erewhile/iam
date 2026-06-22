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
	accessJti := uuid.New()
	refreshJti := uuid.New()

	tokenPair, err := token.Generate(
		userInfo.ID,
		userInfo.UUID,
		sessionID,
		[]byte(config.Get().Token.Aad),
	)
	if err != nil {
		return nil, errors.New("generate token failed")
	}

	txErr := s.transactor.WithTx(ctx, func(txClient *db.Client) error {
		now := utils.Now()

		err := s.token.Create(ctx, txClient, repository.CreateTokenParams{
			UserID:    userInfo.ID,
			JTI:       accessJti,
			SessionID: sessionID,
			Type:      model.TokenTypeAccess,
			TokenHash: hash.HashBlake2b256([]byte(tokenPair.AccessToken)),
			ExpiresAt: now.Add(config.Get().Token.AccessTokenTTL),
			IP:        param.RequestMeta.IP,
			UserAgent: param.RequestMeta.UserAgent,
		})
		if err != nil {
			return err
		}

		return s.token.Create(ctx, txClient, repository.CreateTokenParams{
			UserID:    userInfo.ID,
			JTI:       refreshJti,
			SessionID: sessionID,
			Type:      model.TokenTypeRefresh,
			TokenHash: hash.HashBlake2b256([]byte(tokenPair.RefreshToken)),
			ExpiresAt: now.Add(config.Get().Token.RefreshTokenTTL),
			IP:        param.RequestMeta.IP,
			UserAgent: param.RequestMeta.UserAgent,
		})
	})

	if txErr != nil {
		return nil, fmt.Errorf("save token failed: %w", txErr)
	}

	return tokenPair, nil
}

func (s *UserService) Profile(ctx context.Context, userID int) {}

func (s *UserService) Refresh(ctx context.Context, param req.UserRefresh) {}

func (s *UserService) Logout(ctx context.Context) {}
