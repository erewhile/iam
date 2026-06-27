package service

import (
	"context"
	"errors"

	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/logger"
	"github.com/erewhile/iam/internal/model"
	"github.com/erewhile/iam/internal/repository"
)

type UserRoleService struct {
	repo     repository.UserRoleRepository
	roleRepo repository.RoleRepository
	userRepo repository.UserRepository
}

func NewUserRoleService(
	repo repository.UserRoleRepository,
	roleRepo repository.RoleRepository,
	userRepo repository.UserRepository,
) *UserRoleService {
	return &UserRoleService{
		repo:     repo,
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
}

func (s *UserRoleService) Roles(ctx context.Context, params req.UserRoleRoles) ([]resp.RoleListItem, error) {
	roles, err := s.repo.GetRolesByUserID(ctx, params.UserID)
	if err != nil {
		logger.Error("failed to find roles by user id", err.Error())
		return []resp.RoleListItem{}, errors.New("failed to retrieve roles")
	}

	result := make([]resp.RoleListItem, 0, len(roles))
	for _, role := range roles {
		result = append(result, resp.RoleListItem{
			ID:   role.ID,
			Name: role.Name,
			Code: role.Code,
		})
	}
	return result, nil
}

func (s *UserRoleService) RoleIds(ctx context.Context, params req.UserRoleRoleIds) ([]int, error) {
	roles, err := s.repo.GetRolesByUserID(ctx, params.UserID)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(roles))
	for _, r := range roles {
		ids = append(ids, r.ID)
	}
	return ids, nil
}

func (s *UserRoleService) Assign(ctx context.Context, params req.UserRoleAssignPathParams, body req.UserRoleAssign) error {
	if _, err := s.userRepo.GetByID(ctx, params.UserID); err != nil {
		if db.IsNotFound(err) {
			return errors.New("user not found")
		}
		logger.Error("failed to get user info", err.Error())
		return errors.New("failed to assign roles")
	}

	superAdminID := 0
	superAdminRole, err := s.roleRepo.GetByCode(ctx, model.RoleSuperAdminCode)
	if err != nil {
		if !db.IsNotFound(err) {
			logger.Error("failed to get super admin role", err.Error())
			return errors.New("failed to assign roles")
		}
	} else {
		superAdminID = superAdminRole.ID
	}

	roleIDs := filterAndDedupeRoles(body.RoleIDs, superAdminID)

	if len(roleIDs) > 0 {
		count, err := s.roleRepo.CountByIDs(ctx, roleIDs)
		if err != nil {
			logger.Error("failed to count roles", err.Error())
			return errors.New("failed to assign roles")
		}
		if count != len(roleIDs) {
			return errors.New("role not found")
		}
	}

	if err := s.repo.Replace(ctx, params.UserID, roleIDs); err != nil {
		logger.Error("failed to assign roles", err.Error())
		return errors.New("failed to assign roles")
	}
	return nil
}

func filterAndDedupeRoles(ids []int, excludeID int) []int {
	seen := make(map[int]struct{}, len(ids))
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		if excludeID > 0 && id == excludeID {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
