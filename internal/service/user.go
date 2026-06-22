package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/model"
	"github.com/erewhile/iam/internal/repository"
	"github.com/erewhile/iam/internal/token"
	"github.com/erewhile/iam/pkg/password"
	"github.com/google/uuid"
)

type UserService struct {
	repo  repository.UserRepository
	token repository.TokenRepository
}

func NewUserService(repo repository.UserRepository, token repository.TokenRepository) *UserService {
	return &UserService{repo, token}
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
	tokenPair, err := token.Generate(userInfo.ID, userInfo.UUID, sessionID, []byte(config.Get().Token.Aad))
	if err != nil {
		return nil, errors.New("generate token failed")
	}

	return tokenPair, nil
}

func (s *UserService) Profile(ctx context.Context, userID int) {}

func (s *UserService) Refresh(ctx context.Context, param req.UserRefresh) {}

func (s *UserService) Logout(ctx context.Context) {}
