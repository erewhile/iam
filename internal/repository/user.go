package repository

import (
	"context"

	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/ent/db/user"
	"github.com/erewhile/iam/internal/model"
	"github.com/google/uuid"
)

type CreateUserParams struct {
	Email        string
	Username     string
	PasswordHash []byte
	Status       model.UserStatus
}

type UserRepository interface {
	Create(ctx context.Context, params CreateUserParams) (*db.User, error)
	GetByID(ctx context.Context, id int) (*db.User, error)
	GetByUUID(ctx context.Context, userUUID uuid.UUID) (*db.User, error)
	GetByEmail(ctx context.Context, email string) (*db.User, error)
	GetByUsername(ctx context.Context, username string) (*db.User, error)
}

type userRepository struct {
	*baseRepository
}

var _ UserRepository = (*userRepository)(nil)

func NewUserRepository(client *db.Client) UserRepository {
	return &userRepository{newBaseRepository(client)}
}

func (r *userRepository) Create(ctx context.Context, params CreateUserParams) (*db.User, error) {
	u, err := r.client.User.Create().
		SetEmail(params.Email).
		SetUsername(params.Username).
		SetPasswordHash(params.PasswordHash).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) GetByID(ctx context.Context, userID int) (*db.User, error) {
	u, err := r.client.User.Query().
		Where(user.IDEQ(userID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) GetByUUID(ctx context.Context, userUUID uuid.UUID) (*db.User, error) {
	u, err := r.client.User.Query().
		Where(user.UUIDEQ(userUUID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	u, err := r.client.User.Query().
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*db.User, error) {
	u, err := r.client.User.Query().
		Where(user.UsernameEQ(username)).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}
