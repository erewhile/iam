package service

import (
	"context"
	"errors"

	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/logger"
	"github.com/erewhile/iam/internal/repository"
)

type RoleService struct {
	repo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) *RoleService {
	return &RoleService{repo}
}

func (s *RoleService) List(ctx context.Context, params req.RoleList) ([]resp.RoleListItem, int, error) {
	content, count, err := s.repo.List(ctx, params)
	if err != nil {
		logger.Error("failed to retrieve the list ", err.Error())
		return nil, 0, errors.New("failed to retrieve the list")
	}

	return content, count, nil
}

func (s *RoleService) Info(ctx context.Context, params req.InfoPathParams) (*resp.RoleInfo, error) {
	roleInfo, err := s.repo.GetByID(ctx, params.ID)
	if err != nil {
		logger.Error("failed to get role info", err.Error())
		return nil, errors.New("failed to get role info")
	}

	if roleInfo == nil {
		return nil, errors.New("role not found")
	}

	return &resp.RoleInfo{
		ID:   roleInfo.ID,
		Name: roleInfo.Name,
		Code: roleInfo.Code,
	}, nil
}

func (s *RoleService) Create(ctx context.Context, params req.RoleCreate) error {
	exists, err := s.repo.Duplicate(ctx, params.Name, params.Code)
	if err != nil {
		logger.Error("failed to check if role exists", err)
		return errors.New("failed to check if role exists")
	}

	if exists {
		return errors.New("name or code already exists")
	}

	_, err = s.repo.Create(ctx, params)
	if err != nil {
		logger.Error("failed to create user", err)
		return errors.New("failed to create user")
	}
	return nil
}

func (s *RoleService) Update(ctx context.Context, pathParams req.RoleUpdatePathParams, params req.RoleUpdate) error {
	roleInfo, err := s.repo.GetByID(ctx, pathParams.ID)
	if err != nil {
		logger.Error("get role failed", err.Error())
		return errors.New("failed to get role info")
	}
	if roleInfo == nil {
		return errors.New("user not found")
	}

	exists, err := s.repo.Duplicate(ctx, params.Name, params.Code, pathParams.ID)
	if err != nil {
		logger.Error("failed to check if role exists", err)
		return errors.New("failed to check if role exists")
	}
	if exists {
		return errors.New("name or code already exists")
	}

	_, err = s.repo.Update(ctx, pathParams, params)
	if err != nil {
		logger.Error("failed to update role", err)
		return errors.New("failed to update role")
	}

	return nil
}

func (s *RoleService) Delete(ctx context.Context, pathParams req.DeletePathParams) error {
	roleInfo, err := s.repo.GetByID(ctx, pathParams.ID)
	if err != nil {
		logger.Error("get role failed", err.Error())
		return errors.New("failed to get role info")
	}
	if roleInfo == nil {
		return errors.New("role not found")
	}

	if err := s.repo.Delete(ctx, pathParams); err != nil {
		logger.Error("failed to delete role", err)
		return errors.New("failed to delete role")
	}
	return nil
}
