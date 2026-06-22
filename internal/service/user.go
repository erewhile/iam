package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/model"
	"github.com/erewhile/iam/internal/repository"
	"github.com/erewhile/iam/pkg/password"
)

type UserService struct {
	repo  repository.UserRepository
	token repository.TokenRepository
}

func NewUserService(repo repository.UserRepository, token repository.TokenRepository) *UserService {
	return &UserService{repo, token}
}

func (s *UserService) Login(ctx context.Context, param req.UserLogin) error {
	userInfo, err := s.repo.GetByUsername(ctx, param.Username)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if userInfo == nil {
		return errors.New("user not found")
	}

	if userInfo.Status != model.UserStatusActive {
		return errors.New("account is disabled")
	}

	ok, err := password.Validate(param.Password, string(userInfo.PasswordHash))
	if err != nil {
		return errors.New("password check failed, please try again later")
	}

	if !ok {
		return errors.New("wrong password")
	}

	return nil
}

func (s *UserService) Profile(ctx context.Context, userID int) {}

func (s *UserService) Refresh(ctx context.Context, param req.UserRefresh) {}

func (s *UserService) Logout(ctx context.Context) {}
