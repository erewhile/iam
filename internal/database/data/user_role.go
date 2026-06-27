package data

import (
	"context"
	"fmt"
	"time"

	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/ent/db/role"
	"github.com/erewhile/iam/internal/ent/db/userrole"
	"github.com/erewhile/iam/internal/model"
)

func initUserRole(client *db.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleInfo, err := client.Role.Query().
		Where(role.CodeEQ(model.RoleSuperAdminCode)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("get super admin role failed: %w", err)
	}

	exists, err := client.UserRole.Query().
		Where(
			userrole.UserID(model.UserSystemID),
			userrole.RoleID(roleInfo.ID),
		).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("check user role binding failed: %w", err)
	}
	if exists {
		return nil
	}

	_, err = client.UserRole.Create().
		SetUserID(model.UserSystemID).
		SetRoleID(roleInfo.ID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("bind super admin role to system user failed: %w", err)
	}

	return nil
}
