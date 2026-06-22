package service

import "github.com/erewhile/iam/internal/repository"

type RoleService struct {
	repo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) *RoleService {
	return &RoleService{repo}
}
