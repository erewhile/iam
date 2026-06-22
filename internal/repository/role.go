package repository

import (
	"context"

	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/ent/db/role"
)

type CreateRoleParams struct {
	Code string
	Name string
}

type UpdateRoleParams struct {
	ID   int
	Name string
}

type RoleRepository interface {
	Create(ctx context.Context, params CreateRoleParams) (*db.Role, error)
	Update(ctx context.Context, params UpdateRoleParams) (*db.Role, error)
	Delete(ctx context.Context, id int) error
	FindByID(ctx context.Context, id int) (*db.Role, error)
	FindByCode(ctx context.Context, code string) (*db.Role, error)
	List(ctx context.Context) ([]*db.Role, error)
}

type roleRepository struct {
	*baseRepository
}

var _ RoleRepository = (*roleRepository)(nil)

func NewRoleRepository(client *db.Client) RoleRepository {
	return &roleRepository{newBaseRepository(client)}
}

func (r *roleRepository) Create(ctx context.Context, p CreateRoleParams) (*db.Role, error) {
	return r.client.Role.Create().
		SetCode(p.Code).
		SetName(p.Name).
		Save(ctx)
}

func (r *roleRepository) Update(ctx context.Context, p UpdateRoleParams) (*db.Role, error) {
	return r.client.Role.UpdateOneID(p.ID).
		SetName(p.Name).
		Save(ctx)
}

func (r *roleRepository) Delete(ctx context.Context, id int) error {
	return r.client.Role.DeleteOneID(id).Exec(ctx)
}

func (r *roleRepository) FindByID(ctx context.Context, id int) (*db.Role, error) {
	return r.client.Role.Get(ctx, id)
}

func (r *roleRepository) FindByCode(ctx context.Context, code string) (*db.Role, error) {
	return r.client.Role.Query().Where(role.CodeEQ(code)).Only(ctx)
}

func (r *roleRepository) List(ctx context.Context) ([]*db.Role, error) {
	return r.client.Role.Query().All(ctx)
}
