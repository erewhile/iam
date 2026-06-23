package repository

import (
	"context"

	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/ent/db/user"
	"github.com/erewhile/iam/internal/model"
	"github.com/erewhile/iam/pkg/utils"
	"github.com/google/uuid"
)

type UserRepository interface {
	List(ctx context.Context, params req.UserList) ([]resp.UserListItem, int, error)
	GetByID(ctx context.Context, id int) (*db.User, error)
	GetByUUID(ctx context.Context, userUUID uuid.UUID) (*db.User, error)
	GetByEmail(ctx context.Context, email string) (*db.User, error)
	GetByUsername(ctx context.Context, username string) (*db.User, error)
	Duplicate(ctx context.Context, username, email string, id ...int) (bool, error)
	Create(ctx context.Context, params req.UserCreate, hashed string) (*db.User, error)
	Update(ctx context.Context, params req.UserUpdatePathParams, body req.UserUpdate, hashed string) (*db.User, error)
	Delete(ctx context.Context, params req.DeletePathParams) error
}

type userRepository struct {
	*baseRepository
}

var _ UserRepository = (*userRepository)(nil)

func NewUserRepository(client *db.Client) UserRepository {
	return &userRepository{newBaseRepository(client)}
}

func (r *userRepository) List(ctx context.Context, params req.UserList) ([]resp.UserListItem, int, error) {
	q := r.client.User.Query().
		Where(user.DeletedAtIsNil())

	if params.Keyword != "" {
		q = q.Where(
			user.Or(
				user.EmailContainsFold(params.Keyword),
				user.UsernameContainsFold(params.Keyword),
			),
		)
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []resp.UserListItem{}, 0, nil
	}

	totalPages := (total + params.PerPage - 1) / params.PerPage
	if params.Page > totalPages {
		return []resp.UserListItem{}, total, nil
	}

	offset := (params.Page - 1) * params.PerPage

	users, err := q.
		Order(db.Desc(user.FieldID)).
		Offset(offset).
		Limit(params.PerPage).
		All(ctx)
	if err != nil {
		return []resp.UserListItem{}, 0, err
	}

	result := make([]resp.UserListItem, 0, len(users))
	for _, item := range users {
		result = append(result, resp.UserListItem{
			ID:           item.ID,
			UUID:         item.UUID,
			Email:        item.Email,
			Username:     item.Username,
			StatusDetail: item.Status.String(),
		})
	}

	return result, total, nil
}

func (r *userRepository) GetByID(ctx context.Context, userID int) (*db.User, error) {
	userInfo, err := r.client.User.Query().
		Where(
			user.IDEQ(userID),
			user.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (r *userRepository) GetByUUID(ctx context.Context, userUUID uuid.UUID) (*db.User, error) {
	userInfo, err := r.client.User.Query().
		Where(
			user.UUIDEQ(userUUID),
			user.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	userInfo, err := r.client.User.Query().
		Where(
			user.EmailEQ(email),
			user.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*db.User, error) {
	userInfo, err := r.client.User.Query().
		Where(
			user.UsernameEQ(username),
			user.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return userInfo, nil
}

func (r *userRepository) Duplicate(ctx context.Context, username, email string, id ...int) (bool, error) {
	query := r.client.User.Query().
		Where(
			user.Or(
				user.UsernameEQ(username),
				user.EmailEQ(email),
			),
		)

	if len(id) > 0 && id[0] > 0 {
		query = query.Where(user.IDNEQ(id[0]))
	}

	exist, err := query.Exist(ctx)
	if err != nil {
		return false, err
	}

	return exist, nil
}

func (r *userRepository) Create(ctx context.Context, params req.UserCreate, hashed string) (*db.User, error) {
	createRes, err := r.client.User.Create().
		SetEmail(params.Email).
		SetUsername(params.Username).
		SetPasswordHash([]byte(hashed)).
		SetStatus(params.Status).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return createRes, nil
}

func (r *userRepository) Update(ctx context.Context, params req.UserUpdatePathParams, body req.UserUpdate, hashed string) (*db.User, error) {
	builder := r.client.User.UpdateOneID(params.UserID).
		SetEmail(body.Email).
		SetUsername(body.Username).
		SetStatus(body.Status)

	if hashed != "" {
		builder.SetPasswordHash([]byte(hashed))
	}

	u, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) Delete(ctx context.Context, params req.DeletePathParams) error {
	return r.client.User.UpdateOneID(params.ID).
		SetDeletedAt(utils.Now()).
		SetStatus(model.UserStatusDisabled).
		Exec(ctx)
}
