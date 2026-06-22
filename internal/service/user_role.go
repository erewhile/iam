package service

import "github.com/erewhile/iam/internal/repository"

type UserRoleService struct {
	repo repository.UserRoleRepository
}

func NewUserRoleService(repo repository.UserRoleRepository) *UserRoleService {
	return &UserRoleService{repo}
}
