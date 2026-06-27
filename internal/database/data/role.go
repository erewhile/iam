package data

import (
	"context"
	"fmt"
	"time"

	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/ent/db/role"
	"github.com/erewhile/iam/internal/model"
)

func initRole(client *db.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.Role.Query().
		Where(role.CodeEQ(model.RoleSuperAdminCode)).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("check super admin role failed: %w", err)
	}
	if exists {
		return nil
	}

	_, err = client.Role.Create().
		SetCode(model.RoleSuperAdminCode).
		SetName("Super Admin").
		SetIsSystem(model.RoleSystem).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create super admin role failed: %w", err)
	}

	return nil
}
