package repository

import (
	"context"

	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/ent/db"
	"github.com/erewhile/iam/internal/ent/db/application"
	"github.com/erewhile/iam/pkg/hash"
	"github.com/erewhile/iam/pkg/utils"
)

type ApplicationRepository interface {
	List(ctx context.Context, params req.ApplicationList) ([]resp.ApplicationListItem, int, error)
	GetByID(ctx context.Context, id int) (*db.Application, error)
	GetByClientID(ctx context.Context, clientID string) (*db.Application, error)
	Duplicate(ctx context.Context, name, clientID string, id ...int) (bool, error)
	Create(ctx context.Context, body req.ApplicationCreate, clientSecret string) (*db.Application, error)
	Update(ctx context.Context, params req.ApplicationUpdatePathParams, body req.ApplicationUpdate, clientSecret string) (*db.Application, error)
	UpdateSecret(ctx context.Context, id int, clientSecret string) (*db.Application, error)
	Delete(ctx context.Context, params req.DeletePathParams) error
}

type applicationRepository struct {
	*baseRepository
}

var _ ApplicationRepository = (*applicationRepository)(nil)

func NewApplicationRepository(client *db.Client) ApplicationRepository {
	return &applicationRepository{newBaseRepository(client)}
}

func (r *applicationRepository) List(ctx context.Context, params req.ApplicationList) ([]resp.ApplicationListItem, int, error) {
	q := r.client.Application.Query().
		Where(application.DeletedAtIsNil())

	if params.Keyword != "" {
		q = q.Where(
			application.Or(
				application.NameContainsFold(params.Keyword),
			),
		)
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []resp.ApplicationListItem{}, 0, nil
	}

	totalPages := (total + params.PerPage - 1) / params.PerPage
	if params.Page > totalPages {
		return []resp.ApplicationListItem{}, 0, nil
	}

	offset := (params.Page - 1) * params.PerPage

	applications, err := q.
		Order(db.Desc(application.FieldID)).
		Offset(offset).
		Limit(params.PerPage).
		All(ctx)
	if err != nil {
		return []resp.ApplicationListItem{}, 0, err
	}

	result := make([]resp.ApplicationListItem, 0, len(applications))
	for _, item := range applications {
		result = append(result, resp.ApplicationListItem{
			ID:           item.ID,
			Name:         item.Name,
			ClientID:     item.ClientID,
			RedirectUris: item.RedirectUris,
		})
	}

	return result, total, nil
}

func (r *applicationRepository) GetByID(ctx context.Context, id int) (*db.Application, error) {
	roleInfo, err := r.client.Application.Query().
		Where(
			application.IDEQ(id),
			application.DeletedAtIsNil(),
		).
		Only(ctx)

	if err != nil {
		return nil, err
	}

	return roleInfo, nil
}

func (r *applicationRepository) GetByClientID(ctx context.Context, clientID string) (*db.Application, error) {
	applicationInfo, err := r.client.Application.
		Query().
		Where(
			application.ClientID(clientID),
			application.DeletedAtIsNil(),
		).
		Only(ctx)

	if err != nil {
		return nil, err
	}

	return applicationInfo, nil
}

func (r *applicationRepository) Duplicate(ctx context.Context, name, clientID string, id ...int) (bool, error) {
	query := r.client.Application.Query().
		Where(
			application.Or(
				application.NameEQ(name),
				application.ClientIDEQ(clientID),
			),
		)

	if len(id) > 0 && id[0] > 0 {
		query = query.Where(application.IDNEQ(id[0]))
	}

	exist, err := query.Exist(ctx)
	if err != nil {
		return false, err
	}

	return exist, nil
}

func (r *applicationRepository) Create(ctx context.Context, body req.ApplicationCreate, clientSecret string) (*db.Application, error) {
	createRes, err := r.client.Application.Create().
		SetName(body.Name).
		SetClientID(body.ClientID).
		SetClientSecret(hash.HashBlake2b256([]byte(clientSecret))).
		SetRedirectUris(body.RedirectUris).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return createRes, nil
}

func (r *applicationRepository) Update(ctx context.Context, params req.ApplicationUpdatePathParams, body req.ApplicationUpdate, clientSecret string) (*db.Application, error) {
	updateRes, err := r.client.Application.UpdateOneID(params.ApplicationID).
		SetName(body.Name).
		SetClientID(body.ClientID).
		SetClientSecret(hash.HashBlake2b256([]byte(clientSecret))).
		SetRedirectUris(body.RedirectUris).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return updateRes, nil
}

func (r *applicationRepository) UpdateSecret(ctx context.Context, id int, clientSecret string) (*db.Application, error) {
	updateRes, err := r.client.Application.UpdateOneID(id).
		SetClientSecret(hash.HashBlake2b256([]byte(clientSecret))).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return updateRes, nil
}

func (r *applicationRepository) Delete(ctx context.Context, params req.DeletePathParams) error {
	return r.client.Application.UpdateOneID(params.ID).
		SetDeletedAt(utils.Now()).
		Exec(ctx)
}
