package repository

import (
	"context"

	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/ent/db/user"
	"github.com/erewhile/iam/internal/ent/db/userrole"
)

type UserRoleRepository interface {
	Assign(ctx context.Context, userID int, roleIDs []int) error
	Revoke(ctx context.Context, userID int, roleIDs []int) error
	Replace(ctx context.Context, userID int, roleIDs []int) error
	RevokeAllByUserID(ctx context.Context, userID int) error
	RevokeAllByRoleID(ctx context.Context, roleID int) error
	ExistsByRoleID(ctx context.Context, roleID int) (bool, error)
	GetRolesByUserID(ctx context.Context, userID int) ([]*db.Role, error)
	GetUserIDsByRoleID(ctx context.Context, roleID int) ([]int, error)
}

type userRoleRepository struct {
	*baseRepository
}

var _ UserRoleRepository = (*userRoleRepository)(nil)

func NewUserRoleRepository(client *db.Client) UserRoleRepository {
	return &userRoleRepository{newBaseRepository(client)}
}

func (r *userRoleRepository) Assign(ctx context.Context, userID int, roleIDs []int) error {
	if len(roleIDs) == 0 {
		return nil
	}

	builders := make([]*db.UserRoleCreate, len(roleIDs))
	for i, roleID := range roleIDs {
		builders[i] = r.client.UserRole.Create().
			SetUserID(userID).
			SetRoleID(roleID)
	}

	return r.client.UserRole.CreateBulk(builders...).Exec(ctx)
}

func (r *userRoleRepository) Revoke(ctx context.Context, userID int, roleIDs []int) error {
	if len(roleIDs) == 0 {
		return nil
	}
	_, err := r.client.UserRole.Delete().
		Where(
			userrole.UserIDEQ(userID),
			userrole.RoleIDIn(roleIDs...),
		).
		Exec(ctx)
	return err
}

func (r *userRoleRepository) Replace(ctx context.Context, userID int, roleIDs []int) error {
	if _, err := r.client.UserRole.Delete().
		Where(userrole.UserIDEQ(userID)).
		Exec(ctx); err != nil {
		return err
	}

	if len(roleIDs) == 0 {
		return nil
	}

	builders := make([]*db.UserRoleCreate, len(roleIDs))
	for i, roleID := range roleIDs {
		builders[i] = r.client.UserRole.Create().
			SetUserID(userID).
			SetRoleID(roleID)
	}
	if err := r.client.UserRole.CreateBulk(builders...).Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (r *userRoleRepository) RevokeAllByUserID(ctx context.Context, userID int) error {
	_, err := r.client.UserRole.Delete().
		Where(userrole.UserIDEQ(userID)).
		Exec(ctx)
	return err
}

func (r *userRoleRepository) RevokeAllByRoleID(ctx context.Context, roleID int) error {
	_, err := r.client.UserRole.Delete().
		Where(userrole.RoleIDEQ(roleID)).
		Exec(ctx)
	return err
}

func (r *userRoleRepository) ExistsByRoleID(ctx context.Context, roleID int) (bool, error) {
	return r.client.UserRole.Query().
		Where(userrole.RoleIDEQ(roleID)).
		Exist(ctx)
}

func (r *userRoleRepository) GetRolesByUserID(ctx context.Context, userID int) ([]*db.Role, error) {
	return r.client.User.Query().
		Where(user.IDEQ(userID)).
		QueryRoles().
		All(ctx)
}

func (r *userRoleRepository) GetUserIDsByRoleID(ctx context.Context, roleID int) ([]int, error) {
	return r.client.UserRole.Query().
		Where(userrole.RoleIDEQ(roleID)).
		Select(userrole.FieldUserID).
		Ints(ctx)
}
