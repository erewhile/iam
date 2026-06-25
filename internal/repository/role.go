package repository

import (
	"context"
	"errors"

	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/ent/db/role"
	"github.com/erewhile/iam/internal/ent/db/user"
	"github.com/erewhile/iam/internal/model"
	"github.com/erewhile/iam/pkg/utils"
)

type RoleRepository interface {
	List(ctx context.Context, params req.RoleList) ([]resp.RoleListItem, int, error)
	GetByID(ctx context.Context, id int) (*db.Role, error)
	GetByCode(ctx context.Context, code string) (*db.Role, error)
	ListByUserID(ctx context.Context, userID int) ([]*db.Role, error)
	UserHasRole(ctx context.Context, userID int, code string) (bool, error)
	CountUsersByRoleCode(ctx context.Context, code string) (int, error)
	Duplicate(ctx context.Context, name, code string, id ...int) (bool, error)
	CountByIDs(ctx context.Context, ids []int) (int, error)
	Create(ctx context.Context, body req.RoleCreate) (*db.Role, error)
	Update(ctx context.Context, params req.RoleUpdatePathParams, body req.RoleUpdate) (*db.Role, error)
	Delete(ctx context.Context, params req.DeletePathParams) error
}

type roleRepository struct {
	*baseRepository
}

var _ RoleRepository = (*roleRepository)(nil)

func NewRoleRepository(client *db.Client) RoleRepository {
	return &roleRepository{newBaseRepository(client)}
}

func (r *roleRepository) List(ctx context.Context, params req.RoleList) ([]resp.RoleListItem, int, error) {
	q := r.client.Role.Query().
		Where(role.DeletedAtIsNil())

	if params.Keyword != "" {
		q = q.Where(
			role.Or(
				role.NameContainsFold(params.Keyword),
				role.CodeContainsFold(params.Keyword),
			),
		)
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []resp.RoleListItem{}, 0, nil
	}

	totalPages := (total + params.PerPage - 1) / params.PerPage
	if params.Page > totalPages {
		return []resp.RoleListItem{}, 0, nil
	}

	offset := (params.Page - 1) * params.PerPage

	roles, err := q.
		Order(db.Desc(role.FieldID)).
		Offset(offset).
		Limit(params.PerPage).
		All(ctx)
	if err != nil {
		return []resp.RoleListItem{}, 0, err
	}

	result := make([]resp.RoleListItem, 0, len(roles))
	for _, item := range roles {
		result = append(result, resp.RoleListItem{
			ID:   item.ID,
			Name: item.Name,
			Code: item.Code,
		})
	}

	return result, total, nil
}

func (r *roleRepository) GetByID(ctx context.Context, id int) (*db.Role, error) {
	roleInfo, err := r.client.Role.Query().
		Where(
			role.IDEQ(id),
			role.DeletedAtIsNil(),
		).
		Only(ctx)

	if err != nil {
		return nil, err
	}

	return roleInfo, nil
}

func (r *roleRepository) GetByCode(ctx context.Context, code string) (*db.Role, error) {
	roleInfo, err := r.client.Role.Query().
		Where(
			role.CodeEQ(code),
			role.DeletedAtIsNil(),
		).
		Only(ctx)

	if err != nil {
		return nil, err
	}

	return roleInfo, nil
}

func (r *roleRepository) ListByUserID(ctx context.Context, userID int) ([]*db.Role, error) {
	return r.client.Role.Query().
		Where(
			role.DeletedAtIsNil(),
			role.HasUsersWith(user.IDEQ(userID)),
		).
		All(ctx)
}

func (r *roleRepository) UserHasRole(ctx context.Context, userID int, code string) (bool, error) {
	return r.client.Role.Query().
		Where(
			role.CodeEQ(code),
			role.DeletedAtIsNil(),
			role.HasUsersWith(user.IDEQ(userID)),
		).
		Exist(ctx)
}

func (r *roleRepository) CountUsersByRoleCode(ctx context.Context, code string) (int, error) {
	roleInfo, err := r.GetByCode(ctx, code)
	if err != nil {
		if db.IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}

	return r.client.User.Query().
		Where(
			user.DeletedAtIsNil(),
			user.HasRolesWith(role.IDEQ(roleInfo.ID)),
		).
		Count(ctx)
}

func (r *roleRepository) CountByIDs(ctx context.Context, ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	return r.client.Role.Query().
		Where(
			role.IDIn(ids...),
			role.DeletedAtIsNil(),
		).
		Count(ctx)
}

func (r *roleRepository) Duplicate(ctx context.Context, name, code string, id ...int) (bool, error) {
	query := r.client.Role.Query().
		Where(
			role.Or(
				role.NameEQ(name),
				role.CodeEQ(code),
			),
		)

	if len(id) > 0 && id[0] > 0 {
		query = query.Where(role.IDNEQ(id[0]))
	}

	exist, err := query.Exist(ctx)
	if err != nil {
		return false, err
	}

	return exist, nil
}

func (r *roleRepository) Create(ctx context.Context, body req.RoleCreate) (*db.Role, error) {
	createRes, err := r.client.Role.Create().
		SetCode(body.Code).
		SetName(body.Name).
		SetIsSystem(model.RoleStandard).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return createRes, nil
}

func (r *roleRepository) Update(ctx context.Context, params req.RoleUpdatePathParams, body req.RoleUpdate) (*db.Role, error) {
	updateRes, err := r.client.Role.UpdateOneID(params.RoleID).
		SetName(body.Name).
		SetCode(body.Code).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return updateRes, nil
}

func (r *roleRepository) Delete(ctx context.Context, params req.DeletePathParams) error {
	roleInfo, err := r.GetByID(ctx, params.ID)
	if err != nil {
		return err
	}
	if roleInfo.IsSystem {
		return errors.New("system role cannot be deleted")
	}
	return r.client.Role.UpdateOneID(params.ID).
		SetDeletedAt(utils.Now()).
		Exec(ctx)
}
